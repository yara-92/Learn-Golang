// Package seed 在首次启动、数据库为空时写入一批演示数据：几个不同角色的账号，
// 以及一个能体现"审批人+并行分支+汇合"完整能力的示例模板——报销审批流程：
//
//	   ┌──────────┐
//	   │  start   │
//	   └────┬─────┘
//	        │
//	┌───────▼────────┐
//	│ manager_review │  经理审批（ANY：manager 角色任一人通过）
//	└───────┬────────┘
//	        │
//	┌───────┴────────┐
//	│                 │
//
// ┌───▼────┐      ┌─────▼─────┐
// │hr_review│      │finance_review│   两个分支并行（fan-out）
// └───┬────┘      └─────┬─────┘
//
//	│                 │
//	└───────┬─────────┘
//	        │
//	    ┌───▼───┐
//	    │  end   │   join_type=ALL：必须两个分支都通过才算结束（fan-in）
//	    └───────┘
package seed

import (
	"context"
	"database/sql"
	"log"

	"github.com/yorkyu/approval-engine/internal/auth"
	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
)

// Run 在用户表为空的情况下执行播种；如果已经有数据（比如你重启了服务），
// 则什么都不做，保证幂等、可以放心重复启动。
func Run(ctx context.Context, db *sql.DB) error {
	users, err := repository.ListAllUsers(ctx, db)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		log.Println("[seed] 数据库已有数据，跳过播种")
		return nil
	}

	log.Println("[seed] 首次启动，写入演示账号与示例模板 ...")

	type seedUser struct {
		username, password, display, role string
	}
	demoUsers := []seedUser{
		{"admin", "admin123", "系统管理员", "admin"},
		{"alice", "alice123", "Alice（员工）", "employee"},
		{"bob", "bob123", "Bob（部门经理）", "manager"},
		{"carol", "carol123", "Carol（HR）", "hr"},
		{"dave", "dave123", "Dave（财务）", "finance"},
	}
	ids := map[string]int64{}
	for _, du := range demoUsers {
		hash, err := auth.HashPassword(du.password)
		if err != nil {
			return err
		}
		id, err := repository.CreateUser(ctx, db, &model.User{
			Username: du.username, PasswordHash: hash, DisplayName: du.display, Role: du.role,
		})
		if err != nil {
			return err
		}
		ids[du.username] = id
	}

	def := &model.TemplateDef{
		Name:         "报销审批流程",
		Description:  "员工提交报销 -> 经理审批 -> HR 与财务并行复核（两者都通过才算完成）",
		BusinessType: "EXPENSE",
		Nodes: []model.NodeDef{
			{Code: "start", Name: "开始", NodeType: model.NodeTypeStart},
			{
				Code: "manager_review", Name: "部门经理审批", NodeType: model.NodeTypeApproval,
				ApproveType: model.ApproveTypeAny, JoinType: model.JoinTypeAny,
				Approvers: []model.ApproverDef{{ApproverType: model.ApproverTypeRole, ApproverRef: "manager"}},
			},
			{
				Code: "hr_review", Name: "HR 复核", NodeType: model.NodeTypeApproval,
				ApproveType: model.ApproveTypeAny, JoinType: model.JoinTypeAny,
				Approvers: []model.ApproverDef{{ApproverType: model.ApproverTypeRole, ApproverRef: "hr"}},
			},
			{
				Code: "finance_review", Name: "财务复核", NodeType: model.NodeTypeApproval,
				ApproveType: model.ApproveTypeAny, JoinType: model.JoinTypeAny,
				Approvers: []model.ApproverDef{{ApproverType: model.ApproverTypeRole, ApproverRef: "finance"}},
			},
			// end 的 join_type=ALL：hr_review 和 finance_review 两个分支都
			// 完成才会真正结束流程，这就是"汇合"（fan-in）语义。
			{Code: "end", Name: "结束", NodeType: model.NodeTypeEnd, JoinType: model.JoinTypeAll},
		},
		Edges: []model.EdgeDef{
			{FromCode: "start", ToCode: "manager_review"},
			{FromCode: "manager_review", ToCode: "hr_review"},
			{FromCode: "manager_review", ToCode: "finance_review"},
			{FromCode: "hr_review", ToCode: "end"},
			{FromCode: "finance_review", ToCode: "end"},
		},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := repository.CreateTemplate(ctx, tx, def); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("[seed] 完成。演示账号（用户名/密码）：")
	log.Println("       admin/admin123  bob(经理)/bob123  carol(HR)/carol123  dave(财务)/dave123  alice(员工)/alice123")
	return nil
}
