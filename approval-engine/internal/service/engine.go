// Package service：engine.go 是整个项目的核心 —— 一个基于 DAG（有向无环图）
// 的审批流引擎。设计对应关系（与你之前 Vue+Supabase 版本的概念一一对应）：
//
//	Supabase 版本                          Go 版本
//	------------------------------------   ------------------------------------
//	workflow_templates / nodes / edges     model.TemplateDef -> repository 建表
//	security definer RPC 函数              service.Engine 的方法（业务规则在这里，
//	                                        而不是散落在 handler 里）
//	FOR UPDATE 行锁防止并发重复审批         SQLite 场景下用 engine 级别的 sync.Mutex
//	                                        串行化写路径（下方有详细说明）
//	DAG 模板 + 汇合（fan-in）逻辑           advance() 里的 join_type=ALL 处理
//
// ============================================================================
// 关于并发安全 —— 为什么这里用 sync.Mutex 而不是 SELECT ... FOR UPDATE
// ============================================================================
// SQLite 是"单写者"模型：同一时刻整个数据库只能有一个写事务在进行，这是
// SQLite 自身的机制，不需要也无法使用 Postgres 那种细粒度行锁。
// 我们在 store.Open 里把连接池 SetMaxOpenConns(1)，已经保证了写操作天然串行。
//
// 但只做到"串行"还不够：ApproveTask 内部是"读任务状态 -> 判断 -> 写状态"的
// 复合操作，如果两个 goroutine 同时读到"任务是 PENDING"，都会尝试推进流程，
// 造成同一个节点被重复通过。所以我们在 service 层再加一把进程内 sync.Mutex，
// 把"读-判断-写"这个复合操作当成一个临界区。
//
// 迁移到 Postgres 生产环境时的正确做法：把这里的 Mutex 去掉，改为在事务里
// `SELECT * FROM workflow_tasks WHERE id = ? FOR UPDATE`，让数据库本身的
// 行锁来保证同一任务不会被并发处理——这也是多实例水平扩展时唯一正确的方案
// （进程内 Mutex 在多实例部署下不会跨进程生效）。
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
)

type Engine struct {
	db *sql.DB
	mu sync.Mutex // 见文件头注释：串行化"读状态->判断->写状态"的复合操作
}

func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// templateGraph 是模板在内存中的图结构，每次引擎操作时从数据库加载一次
// （对审批系统而言模板数据量小、变更不频繁，没必要额外做缓存失效的复杂度）。
type templateGraph struct {
	tpl        *model.Template
	nodeByID   map[int64]*model.Node
	nodeByCode map[string]*model.Node
	outEdges   map[int64][]int64 // fromNodeID -> []toNodeID
	inEdges    map[int64][]int64 // toNodeID -> []fromNodeID
}

func loadGraph(ctx context.Context, q repository.Querier, templateID int64) (*templateGraph, error) {
	tpl, err := repository.GetTemplate(ctx, q, templateID)
	if err != nil {
		return nil, err
	}
	g := &templateGraph{
		tpl:        tpl,
		nodeByID:   map[int64]*model.Node{},
		nodeByCode: map[string]*model.Node{},
		outEdges:   map[int64][]int64{},
		inEdges:    map[int64][]int64{},
	}
	for i := range tpl.Nodes {
		n := &tpl.Nodes[i]
		g.nodeByID[n.ID] = n
		g.nodeByCode[n.Code] = n
	}
	for _, e := range tpl.Edges {
		g.outEdges[e.FromNodeID] = append(g.outEdges[e.FromNodeID], e.ToNodeID)
		g.inEdges[e.ToNodeID] = append(g.inEdges[e.ToNodeID], e.FromNodeID)
	}
	return g, nil
}

// ---------------------------------------------------------------------------
// StartInstance：发起一个审批实例
// ---------------------------------------------------------------------------

type StartInstanceInput struct {
	TemplateID   int64
	BusinessType string
	BusinessID   string
	Title        string
	FormData     map[string]any
	InitiatorID  int64
}

func (e *Engine) StartInstance(ctx context.Context, in StartInstanceInput) (*model.Instance, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // commit 成功后 Rollback 是 no-op，这是标准写法

	graph, err := loadGraph(ctx, tx, in.TemplateID)
	if err != nil {
		return nil, err
	}
	startNode := findNodeByType(graph, model.NodeTypeStart)
	if startNode == nil {
		return nil, fmt.Errorf("%w: template %d has no START node", model.ErrTemplateInvalid, in.TemplateID)
	}

	formDataJSON, err := json.Marshal(in.FormData)
	if err != nil {
		return nil, err
	}

	instance := &model.Instance{
		TemplateID:   in.TemplateID,
		BusinessType: in.BusinessType,
		BusinessID:   in.BusinessID,
		Title:        in.Title,
		FormData:     string(formDataJSON),
		InitiatorID:  in.InitiatorID,
		Status:       model.InstanceStatusRunning,
	}
	instanceID, err := repository.CreateInstance(ctx, tx, instance)
	if err != nil {
		return nil, err
	}
	instance.ID = instanceID

	if err := repository.CreateLog(ctx, tx, instanceID, &in.InitiatorID, "START", "发起审批流程"); err != nil {
		return nil, err
	}

	// START 节点无需审批人，直接标记为 APPROVED，然后向下推进。
	startInstNodeID, err := repository.CreateInstanceNode(ctx, tx, instanceID, startNode.ID, model.InstNodeStatusApproved)
	if err != nil {
		return nil, err
	}
	_ = startInstNodeID

	if err := e.advance(ctx, tx, graph, instance, startNode); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return instance, nil
}

// ---------------------------------------------------------------------------
// ApproveTask / RejectTask：审批人对任务做出决定
// ---------------------------------------------------------------------------

func (e *Engine) ApproveTask(ctx context.Context, taskID, actorID int64, comment string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	task, err := repository.GetTaskByID(ctx, tx, taskID)
	if err != nil {
		return err
	}
	if task.ApproverID != actorID {
		return model.ErrForbidden
	}
	if task.Status != model.TaskStatusPending {
		return model.ErrTaskNotPending
	}

	if err := repository.UpdateTaskResult(ctx, tx, taskID, model.TaskStatusApproved, comment); err != nil {
		return err
	}
	if err := repository.CreateLog(ctx, tx, task.InstanceID, &actorID, "APPROVE", comment); err != nil {
		return err
	}

	instNode, err := repository.GetInstanceNodeByID(ctx, tx, task.InstanceNodeID)
	if err != nil {
		return err
	}
	instance, err := repository.GetInstance(ctx, tx, task.InstanceID)
	if err != nil {
		return err
	}
	graph, err := loadGraph(ctx, tx, instance.TemplateID)
	if err != nil {
		return err
	}
	node := graph.nodeByID[instNode.NodeID]

	nodePassed := true
	if node.ApproveType == model.ApproveTypeAll {
		pending, err := repository.CountPendingTasksInNode(ctx, tx, instNode.ID)
		if err != nil {
			return err
		}
		nodePassed = pending == 0
	} else {
		// ANY：一人通过即通过，其余候选人的任务不再需要处理
		if err := repository.CancelPendingTasksInNode(ctx, tx, instNode.ID, taskID); err != nil {
			return err
		}
	}

	if nodePassed {
		if err := repository.UpdateInstanceNodeStatus(ctx, tx, instNode.ID, model.InstNodeStatusApproved, true); err != nil {
			return err
		}
		if err := e.advance(ctx, tx, graph, instance, node); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *Engine) RejectTask(ctx context.Context, taskID, actorID int64, comment string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	task, err := repository.GetTaskByID(ctx, tx, taskID)
	if err != nil {
		return err
	}
	if task.ApproverID != actorID {
		return model.ErrForbidden
	}
	if task.Status != model.TaskStatusPending {
		return model.ErrTaskNotPending
	}

	if err := repository.UpdateTaskResult(ctx, tx, taskID, model.TaskStatusRejected, comment); err != nil {
		return err
	}
	if err := repository.CreateLog(ctx, tx, task.InstanceID, &actorID, "REJECT", comment); err != nil {
		return err
	}
	if err := repository.UpdateInstanceNodeStatus(ctx, tx, task.InstanceNodeID, model.InstNodeStatusRejected, true); err != nil {
		return err
	}
	// 一票否决：整条流程终止，取消其余所有未处理任务。
	if err := repository.CancelAllPendingTasksForInstance(ctx, tx, task.InstanceID); err != nil {
		return err
	}
	if err := repository.UpdateInstanceStatus(ctx, tx, task.InstanceID, model.InstanceStatusRejected); err != nil {
		return err
	}
	if err := repository.CreateLog(ctx, tx, task.InstanceID, nil, "INSTANCE_REJECTED", "流程被拒绝，已终止"); err != nil {
		return err
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// advance：某个节点通过后，沿着 DAG 的边向后推进，处理 fan-out / fan-in
// ---------------------------------------------------------------------------

func (e *Engine) advance(ctx context.Context, tx *sql.Tx, graph *templateGraph, instance *model.Instance, fromNode *model.Node) error {
	targets := graph.outEdges[fromNode.ID]

	if len(targets) == 0 {
		// 没有出边：这是一个"死路"节点（模板设计问题，或者就是有意的终点），
		// 保守起见直接把实例标记为通过。
		return e.completeInstance(ctx, tx, instance, model.InstanceStatusApproved)
	}

	for _, targetID := range targets {
		target := graph.nodeByID[targetID]

		if target.NodeType == model.NodeTypeEnd {
			if err := e.tryActivateEnd(ctx, tx, graph, instance, target); err != nil {
				return err
			}
			continue
		}

		if target.JoinType == model.JoinTypeAll {
			sourceIDs := graph.inEdges[target.ID]
			approvedCount, err := repository.CountApprovedInstanceNodes(ctx, tx, instance.ID, sourceIDs)
			if err != nil {
				return err
			}
			if approvedCount < len(sourceIDs) {
				// 还有兄弟分支没通过，先不激活，等最后一个分支通过时会再次
				// 触发 advance 并重新统计。
				continue
			}
		}

		already, err := repository.GetInstanceNodeByNodeID(ctx, tx, instance.ID, target.ID)
		if err != nil && !errors.Is(err, model.ErrNotFound) {
			return err
		}
		if already != nil {
			// 已经激活/完成过（例如 ANY 汇合下多个分支都到达同一节点），跳过。
			continue
		}

		if err := e.activateNode(ctx, tx, instance, target); err != nil {
			return err
		}
	}
	return nil
}

// tryActivateEnd 处理指向 END 节点的边：END 节点同样支持 join_type=ALL，
// 意味着"必须所有分支都完成才算流程结束"。
func (e *Engine) tryActivateEnd(ctx context.Context, tx *sql.Tx, graph *templateGraph, instance *model.Instance, end *model.Node) error {
	if end.JoinType == model.JoinTypeAll {
		sourceIDs := graph.inEdges[end.ID]
		approvedCount, err := repository.CountApprovedInstanceNodes(ctx, tx, instance.ID, sourceIDs)
		if err != nil {
			return err
		}
		if approvedCount < len(sourceIDs) {
			return nil // 还有分支没完成，等下一次触发
		}
	}
	return e.completeInstance(ctx, tx, instance, model.InstanceStatusApproved)
}

func (e *Engine) completeInstance(ctx context.Context, tx *sql.Tx, instance *model.Instance, status string) error {
	if err := repository.UpdateInstanceStatus(ctx, tx, instance.ID, status); err != nil {
		return err
	}
	return repository.CreateLog(ctx, tx, instance.ID, nil, "INSTANCE_"+status, "流程结束")
}

// activateNode 激活一个审批节点：创建 instance_node，解析审批人（USER / ROLE），
// 为每个审批人生成待办任务。
func (e *Engine) activateNode(ctx context.Context, tx *sql.Tx, instance *model.Instance, node *model.Node) error {
	instNodeID, err := repository.CreateInstanceNode(ctx, tx, instance.ID, node.ID, model.InstNodeStatusActive)
	if err != nil {
		return err
	}

	approverIDs, err := resolveApprovers(ctx, tx, node.Approvers)
	if err != nil {
		return err
	}
	if len(approverIDs) == 0 {
		return fmt.Errorf("%w: node %q has no resolvable approvers", model.ErrTemplateInvalid, node.Code)
	}

	for _, uid := range approverIDs {
		if _, err := repository.CreateTask(ctx, tx, instance.ID, instNodeID, uid); err != nil {
			return err
		}
	}

	detail := fmt.Sprintf("进入节点【%s】，待审批人数：%d", node.Name, len(approverIDs))
	return repository.CreateLog(ctx, tx, instance.ID, nil, "NODE_ACTIVATED", detail)
}

// resolveApprovers 把节点定义里的 USER/ROLE 审批人配置展开成具体的用户 ID 列表，
// 并去重（同一个人可能既被指定为具名审批人、又命中角色审批人）。
func resolveApprovers(ctx context.Context, q repository.Querier, defs []model.NodeApprover) ([]int64, error) {
	seen := map[int64]bool{}
	var out []int64
	for _, d := range defs {
		switch d.ApproverType {
		case model.ApproverTypeUser:
			var uid int64
			if _, err := fmt.Sscanf(d.ApproverRef, "%d", &uid); err != nil {
				return nil, fmt.Errorf("invalid user approver ref %q: %w", d.ApproverRef, err)
			}
			if !seen[uid] {
				seen[uid] = true
				out = append(out, uid)
			}
		case model.ApproverTypeRole:
			users, err := repository.ListUsersByRole(ctx, q, d.ApproverRef)
			if err != nil {
				return nil, err
			}
			for _, u := range users {
				if !seen[u.ID] {
					seen[u.ID] = true
					out = append(out, u.ID)
				}
			}
		}
	}
	return out, nil
}

func findNodeByType(g *templateGraph, nodeType string) *model.Node {
	for _, n := range g.tpl.Nodes {
		if n.NodeType == nodeType {
			return g.nodeByID[n.ID]
		}
	}
	return nil
}
