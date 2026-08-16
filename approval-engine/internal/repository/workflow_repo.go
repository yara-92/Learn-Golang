package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yorkyu/approval-engine/internal/model"
)

// CreateTemplate 在一个事务内创建模板 + 节点 + 节点审批人 + 边。
// 调用方（service 层）负责传入一个 *sql.Tx 作为 q，保证整体原子性。
func CreateTemplate(ctx context.Context, q Querier, def *model.TemplateDef) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO workflow_templates (name, description, business_type, is_active) VALUES (?, ?, ?, 1)`,
		def.Name, def.Description, def.BusinessType,
	)
	if err != nil {
		return 0, err
	}
	templateID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// code -> 数据库自增 id 的映射，用于把 Edge 里引用的 code 转成真实 node_id
	codeToID := make(map[string]int64, len(def.Nodes))

	for _, n := range def.Nodes {
		nres, err := q.ExecContext(ctx,
			`INSERT INTO workflow_nodes (template_id, code, name, node_type, approve_type, join_type) VALUES (?, ?, ?, ?, ?, ?)`,
			templateID, n.Code, n.Name, n.NodeType, n.ApproveType, n.JoinType,
		)
		if err != nil {
			return 0, err
		}
		nodeID, err := nres.LastInsertId()
		if err != nil {
			return 0, err
		}
		codeToID[n.Code] = nodeID

		for _, ap := range n.Approvers {
			if _, err := q.ExecContext(ctx,
				`INSERT INTO workflow_node_approvers (node_id, approver_type, approver_ref) VALUES (?, ?, ?)`,
				nodeID, ap.ApproverType, ap.ApproverRef,
			); err != nil {
				return 0, err
			}
		}
	}

	for _, e := range def.Edges {
		fromID, ok1 := codeToID[e.FromCode]
		toID, ok2 := codeToID[e.ToCode]
		if !ok1 || !ok2 {
			return 0, model.ErrTemplateInvalid
		}
		if _, err := q.ExecContext(ctx,
			`INSERT INTO workflow_edges (template_id, from_node_id, to_node_id) VALUES (?, ?, ?)`,
			templateID, fromID, toID,
		); err != nil {
			return 0, err
		}
	}

	return templateID, nil
}

// GetTemplate 加载模板及其全部节点/审批人/边，用于引擎运行时构图。
func GetTemplate(ctx context.Context, q Querier, id int64) (*model.Template, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, name, description, business_type, is_active, created_at FROM workflow_templates WHERE id = ?`,
		id,
	)
	var t model.Template
	var isActive int
	if err := row.Scan(&t.ID, &t.Name, &t.Description, &t.BusinessType, &isActive, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	t.IsActive = isActive == 1

	nodeRows, err := q.QueryContext(ctx,
		`SELECT id, template_id, code, name, node_type, approve_type, join_type FROM workflow_nodes WHERE template_id = ? ORDER BY id`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer nodeRows.Close()
	for nodeRows.Next() {
		var n model.Node
		if err := nodeRows.Scan(&n.ID, &n.TemplateID, &n.Code, &n.Name, &n.NodeType, &n.ApproveType, &n.JoinType); err != nil {
			return nil, err
		}
		t.Nodes = append(t.Nodes, n)
	}
	if err := nodeRows.Err(); err != nil {
		return nil, err
	}

	for i := range t.Nodes {
		apRows, err := q.QueryContext(ctx,
			`SELECT id, node_id, approver_type, approver_ref FROM workflow_node_approvers WHERE node_id = ? ORDER BY id`,
			t.Nodes[i].ID,
		)
		if err != nil {
			return nil, err
		}
		for apRows.Next() {
			var ap model.NodeApprover
			if err := apRows.Scan(&ap.ID, &ap.NodeID, &ap.ApproverType, &ap.ApproverRef); err != nil {
				apRows.Close()
				return nil, err
			}
			t.Nodes[i].Approvers = append(t.Nodes[i].Approvers, ap)
		}
		apRows.Close()
	}

	edgeRows, err := q.QueryContext(ctx,
		`SELECT id, template_id, from_node_id, to_node_id FROM workflow_edges WHERE template_id = ? ORDER BY id`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer edgeRows.Close()
	for edgeRows.Next() {
		var e model.Edge
		if err := edgeRows.Scan(&e.ID, &e.TemplateID, &e.FromNodeID, &e.ToNodeID); err != nil {
			return nil, err
		}
		t.Edges = append(t.Edges, e)
	}

	return &t, edgeRows.Err()
}

func ListTemplates(ctx context.Context, q Querier) ([]model.Template, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, name, description, business_type, is_active, created_at FROM workflow_templates ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Template
	for rows.Next() {
		var t model.Template
		var isActive int
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.BusinessType, &isActive, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.IsActive = isActive == 1
		out = append(out, t)
	}
	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
