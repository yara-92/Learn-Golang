package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yorkyu/approval-engine/internal/model"
)

// ---------- workflow_instances ----------

func CreateInstance(ctx context.Context, q Querier, in *model.Instance) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO workflow_instances (template_id, business_type, business_id, title, form_data, initiator_id, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.TemplateID, in.BusinessType, in.BusinessID, in.Title, in.FormData, in.InitiatorID, in.Status,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetInstance(ctx context.Context, q Querier, id int64) (*model.Instance, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, template_id, business_type, business_id, title, form_data, initiator_id, status, created_at, updated_at
		 FROM workflow_instances WHERE id = ?`, id)
	var in model.Instance
	err := row.Scan(&in.ID, &in.TemplateID, &in.BusinessType, &in.BusinessID, &in.Title, &in.FormData,
		&in.InitiatorID, &in.Status, &in.CreatedAt, &in.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &in, nil
}

func UpdateInstanceStatus(ctx context.Context, q Querier, id int64, status string) error {
	_, err := q.ExecContext(ctx,
		`UPDATE workflow_instances SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	return err
}

func ListInstancesByInitiator(ctx context.Context, q Querier, initiatorID int64) ([]model.Instance, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, template_id, business_type, business_id, title, form_data, initiator_id, status, created_at, updated_at
		 FROM workflow_instances WHERE initiator_id = ? ORDER BY id DESC`, initiatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInstances(rows)
}

func ListAllInstances(ctx context.Context, q Querier) ([]model.Instance, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, template_id, business_type, business_id, title, form_data, initiator_id, status, created_at, updated_at
		 FROM workflow_instances ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInstances(rows)
}

func scanInstances(rows *sql.Rows) ([]model.Instance, error) {
	var out []model.Instance
	for rows.Next() {
		var in model.Instance
		if err := rows.Scan(&in.ID, &in.TemplateID, &in.BusinessType, &in.BusinessID, &in.Title, &in.FormData,
			&in.InitiatorID, &in.Status, &in.CreatedAt, &in.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, in)
	}
	return out, rows.Err()
}

// ---------- workflow_instance_nodes ----------

func CreateInstanceNode(ctx context.Context, q Querier, instanceID, nodeID int64, status string) (int64, error) {
	now := time.Now()
	var activatedAt any
	if status == model.InstNodeStatusActive || status == model.InstNodeStatusApproved {
		activatedAt = now
	}
	res, err := q.ExecContext(ctx,
		`INSERT INTO workflow_instance_nodes (instance_id, node_id, status, activated_at) VALUES (?, ?, ?, ?)`,
		instanceID, nodeID, status, activatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetInstanceNodeByNodeID(ctx context.Context, q Querier, instanceID, nodeID int64) (*model.InstanceNode, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, instance_id, node_id, status, activated_at, completed_at
		 FROM workflow_instance_nodes WHERE instance_id = ? AND node_id = ?`, instanceID, nodeID)
	return scanInstanceNode(row)
}

func GetInstanceNodeByID(ctx context.Context, q Querier, id int64) (*model.InstanceNode, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, instance_id, node_id, status, activated_at, completed_at
		 FROM workflow_instance_nodes WHERE id = ?`, id)
	return scanInstanceNode(row)
}

func scanInstanceNode(row *sql.Row) (*model.InstanceNode, error) {
	var n model.InstanceNode
	err := row.Scan(&n.ID, &n.InstanceID, &n.NodeID, &n.Status, &n.ActivatedAt, &n.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func UpdateInstanceNodeStatus(ctx context.Context, q Querier, id int64, status string, completed bool) error {
	if completed {
		_, err := q.ExecContext(ctx,
			`UPDATE workflow_instance_nodes SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`,
			status, id)
		return err
	}
	_, err := q.ExecContext(ctx,
		`UPDATE workflow_instance_nodes SET status = ? WHERE id = ?`, status, id)
	return err
}

// CountApprovedInstanceNodes 统计给定 nodeIDs 中，在该 instance 下已经是
// APPROVED 状态的数量——用于 join_type=ALL 的多分支汇合判断。
func CountApprovedInstanceNodes(ctx context.Context, q Querier, instanceID int64, nodeIDs []int64) (int, error) {
	if len(nodeIDs) == 0 {
		return 0, nil
	}
	query := `SELECT COUNT(*) FROM workflow_instance_nodes WHERE instance_id = ? AND status = 'APPROVED' AND node_id IN (` + placeholders(len(nodeIDs)) + `)`
	args := make([]any, 0, len(nodeIDs)+1)
	args = append(args, instanceID)
	for _, id := range nodeIDs {
		args = append(args, id)
	}
	var count int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func ListInstanceNodesByInstance(ctx context.Context, q Querier, instanceID int64) ([]model.InstanceNode, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, instance_id, node_id, status, activated_at, completed_at
		 FROM workflow_instance_nodes WHERE instance_id = ? ORDER BY id`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.InstanceNode
	for rows.Next() {
		var n model.InstanceNode
		if err := rows.Scan(&n.ID, &n.InstanceID, &n.NodeID, &n.Status, &n.ActivatedAt, &n.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// ---------- workflow_tasks ----------

func CreateTask(ctx context.Context, q Querier, instanceID, instanceNodeID, approverID int64) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO workflow_tasks (instance_id, instance_node_id, approver_id, status) VALUES (?, ?, ?, 'PENDING')`,
		instanceID, instanceNodeID, approverID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetTaskByID(ctx context.Context, q Querier, id int64) (*model.Task, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, instance_id, instance_node_id, approver_id, status, comment, acted_at, created_at
		 FROM workflow_tasks WHERE id = ?`, id)
	var t model.Task
	err := row.Scan(&t.ID, &t.InstanceID, &t.InstanceNodeID, &t.ApproverID, &t.Status, &t.Comment, &t.ActedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func UpdateTaskResult(ctx context.Context, q Querier, id int64, status, comment string) error {
	_, err := q.ExecContext(ctx,
		`UPDATE workflow_tasks SET status = ?, comment = ?, acted_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, comment, id)
	return err
}

// CancelPendingTasksInNode 把同一个 instance_node 下其它仍处于 PENDING 的任务取消——
// 用于 approve_type=ANY 时，一人通过后其余候选审批人的任务就不再需要处理。
func CancelPendingTasksInNode(ctx context.Context, q Querier, instanceNodeID, exceptTaskID int64) error {
	_, err := q.ExecContext(ctx,
		`UPDATE workflow_tasks SET status = 'CANCELLED' WHERE instance_node_id = ? AND status = 'PENDING' AND id != ?`,
		instanceNodeID, exceptTaskID)
	return err
}

// CancelAllPendingTasksForInstance 在流程被拒绝/撤销时，取消该实例下所有未处理任务。
func CancelAllPendingTasksForInstance(ctx context.Context, q Querier, instanceID int64) error {
	_, err := q.ExecContext(ctx,
		`UPDATE workflow_tasks SET status = 'CANCELLED' WHERE instance_id = ? AND status = 'PENDING'`,
		instanceID)
	return err
}

func CountPendingTasksInNode(ctx context.Context, q Querier, instanceNodeID int64) (int, error) {
	var count int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM workflow_tasks WHERE instance_node_id = ? AND status = 'PENDING'`,
		instanceNodeID).Scan(&count)
	return count, err
}

func ListTasksByApprover(ctx context.Context, q Querier, approverID int64, status string) ([]model.Task, error) {
	query := `SELECT t.id, t.instance_id, t.instance_node_id, t.approver_id, t.status, t.comment, t.acted_at, t.created_at,
	                  n.name, i.title
	           FROM workflow_tasks t
	           JOIN workflow_instance_nodes inode ON inode.id = t.instance_node_id
	           JOIN workflow_nodes n ON n.id = inode.node_id
	           JOIN workflow_instances i ON i.id = t.instance_id
	           WHERE t.approver_id = ?`
	args := []any{approverID}
	if status != "" {
		query += ` AND t.status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY t.id DESC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.InstanceID, &t.InstanceNodeID, &t.ApproverID, &t.Status, &t.Comment,
			&t.ActedAt, &t.CreatedAt, &t.NodeName, &t.InstanceTitle); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func ListTasksByInstance(ctx context.Context, q Querier, instanceID int64) ([]model.Task, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT t.id, t.instance_id, t.instance_node_id, t.approver_id, t.status, t.comment, t.acted_at, t.created_at, n.name
		 FROM workflow_tasks t
		 JOIN workflow_instance_nodes inode ON inode.id = t.instance_node_id
		 JOIN workflow_nodes n ON n.id = inode.node_id
		 WHERE t.instance_id = ? ORDER BY t.id`, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.InstanceID, &t.InstanceNodeID, &t.ApproverID, &t.Status, &t.Comment,
			&t.ActedAt, &t.CreatedAt, &t.NodeName); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ---------- workflow_logs ----------

func CreateLog(ctx context.Context, q Querier, instanceID int64, actorID *int64, action, detail string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO workflow_logs (instance_id, actor_id, action, detail) VALUES (?, ?, ?, ?)`,
		instanceID, actorID, action, detail)
	return err
}

func ListLogsByInstance(ctx context.Context, q Querier, instanceID int64) ([]model.Log, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, instance_id, actor_id, action, detail, created_at FROM workflow_logs WHERE instance_id = ? ORDER BY id`,
		instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Log
	for rows.Next() {
		var l model.Log
		if err := rows.Scan(&l.ID, &l.InstanceID, &l.ActorID, &l.Action, &l.Detail, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// ---------- helpers ----------

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	s := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			s = append(s, ',')
		}
		s = append(s, '?')
	}
	return string(s)
}
