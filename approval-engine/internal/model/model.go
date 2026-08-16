package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Time 包装 time.Time，用于安全地从 SQLite 驱动读取 DATETIME 列。
// 不同 SQLite 驱动对 DATETIME 的底层返回类型不完全一致（time.Time / string /
// []byte 都有可能），这里统一做兼容解析，同时用 Valid 字段表达 SQL NULL，
// 避免直接用 time.Time 或 *time.Time 在个别驱动实现下 Scan 报错。
type Time struct {
	T     time.Time
	Valid bool
}

func NewTime(t time.Time) Time { return Time{T: t, Valid: true} }

var sqliteTimeLayouts = []string{
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02 15:04:05",
	time.RFC3339Nano,
	time.RFC3339,
}

// Scan 实现 database/sql.Scanner 接口。
func (t *Time) Scan(value any) error {
	if value == nil {
		t.T, t.Valid = time.Time{}, false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.T, t.Valid = v, true
		return nil
	case string:
		return t.parse(v)
	case []byte:
		return t.parse(string(v))
	default:
		return fmt.Errorf("model.Time: unsupported scan type %T", value)
	}
}

func (t *Time) parse(s string) error {
	for _, layout := range sqliteTimeLayouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.T, t.Valid = parsed, true
			return nil
		}
	}
	return fmt.Errorf("model.Time: cannot parse %q", s)
}

// Value 实现 database/sql/driver.Valuer 接口。
func (t Time) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.T, nil
}

// MarshalJSON 让零值/NULL 输出为 JSON null，而不是 "0001-01-01T00:00:00Z"。
func (t Time) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.T)
}

// ---------- 用户 ----------

type User struct {
	ID           int64  `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"password_hash"`
	DisplayName  string `json:"display_name" db:"display_name"`
	Role         string `json:"role" db:"role"` // employee/manager/hr/finance/admin
	CreatedAt    Time   `json:"created_at" db:"created_at"`
}

// ---------- 工作流模板 ----------

const (
	NodeTypeStart    = "START"
	NodeTypeApproval = "APPROVAL"
	NodeTypeEnd      = "END"

	ApproveTypeAny = "ANY" // 节点内多个审批人，任一人通过即通过
	ApproveTypeAll = "ALL" // 节点内多个审批人，需全部通过

	JoinTypeAny = "ANY" // 多条入边，任一分支到达即可激活
	JoinTypeAll = "ALL" // 多条入边，需全部分支都通过才能激活（fan-in 汇合）

	ApproverTypeUser = "USER"
	ApproverTypeRole = "ROLE"
)

type Template struct {
	ID           int64  `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	BusinessType string `json:"business_type" db:"business_type"`
	IsActive     bool   `json:"is_active" db:"is_active"`
	CreatedAt    Time   `json:"created_at" db:"created_at"`

	Nodes []Node `json:"nodes,omitempty" db:"-"`
	Edges []Edge `json:"edges,omitempty" db:"-"`
}

type Node struct {
	ID          int64  `json:"id" db:"id"`
	TemplateID  int64  `json:"template_id" db:"template_id"`
	Code        string `json:"code" db:"code"`
	Name        string `json:"name" db:"name"`
	NodeType    string `json:"node_type" db:"node_type"`
	ApproveType string `json:"approve_type" db:"approve_type"`
	JoinType    string `json:"join_type" db:"join_type"`

	Approvers []NodeApprover `json:"approvers,omitempty" db:"-"`
}

type NodeApprover struct {
	ID           int64  `json:"id" db:"id"`
	NodeID       int64  `json:"node_id" db:"node_id"`
	ApproverType string `json:"approver_type" db:"approver_type"` // USER / ROLE
	ApproverRef  string `json:"approver_ref" db:"approver_ref"`   // user_id 或 role 名
}

type Edge struct {
	ID         int64 `json:"id" db:"id"`
	TemplateID int64 `json:"template_id" db:"template_id"`
	FromNodeID int64 `json:"from_node_id" db:"from_node_id"`
	ToNodeID   int64 `json:"to_node_id" db:"to_node_id"`
}

// ---------- 模板"定义"输入（用于创建模板，用 Code 而非数据库自增 ID 引用节点）----------
//
// 创建模板时节点还没有落库、拿不到自增 ID，所以边只能先用业务方自定义的
// Code（如 "start"/"manager"/"end"）来表达"从哪个节点到哪个节点"，
// repository 层在插入节点后会把 Code 换算成真实的 node_id。

type TemplateDef struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	BusinessType string    `json:"business_type"`
	Nodes        []NodeDef `json:"nodes"`
	Edges        []EdgeDef `json:"edges"`
}

type NodeDef struct {
	Code        string        `json:"code"`
	Name        string        `json:"name"`
	NodeType    string        `json:"node_type"`    // START / APPROVAL / END
	ApproveType string        `json:"approve_type"` // ANY / ALL，默认 ANY
	JoinType    string        `json:"join_type"`    // ANY / ALL，默认 ANY
	Approvers   []ApproverDef `json:"approvers"`
}

type ApproverDef struct {
	ApproverType string `json:"approver_type"` // USER / ROLE
	ApproverRef  string `json:"approver_ref"`  // 用户 ID（字符串形式）或角色名
}

type EdgeDef struct {
	FromCode string `json:"from_code"`
	ToCode   string `json:"to_code"`
}

// ---------- 工作流实例 ----------

const (
	InstanceStatusRunning  = "RUNNING"
	InstanceStatusApproved = "APPROVED"
	InstanceStatusRejected = "REJECTED"
	InstanceStatusCanceled = "CANCELLED"

	InstNodeStatusPending  = "PENDING"
	InstNodeStatusActive   = "ACTIVE"
	InstNodeStatusApproved = "APPROVED"
	InstNodeStatusRejected = "REJECTED"
	InstNodeStatusCanceled = "CANCELLED"

	TaskStatusPending  = "PENDING"
	TaskStatusApproved = "APPROVED"
	TaskStatusRejected = "REJECTED"
	TaskStatusCanceled = "CANCELLED"
)

type Instance struct {
	ID           int64  `json:"id" db:"id"`
	TemplateID   int64  `json:"template_id" db:"template_id"`
	BusinessType string `json:"business_type" db:"business_type"`
	BusinessID   string `json:"business_id" db:"business_id"`
	Title        string `json:"title" db:"title"`
	FormData     string `json:"form_data" db:"form_data"` // JSON 字符串
	InitiatorID  int64  `json:"initiator_id" db:"initiator_id"`
	Status       string `json:"status" db:"status"`
	CreatedAt    Time   `json:"created_at" db:"created_at"`
	UpdatedAt    Time   `json:"updated_at" db:"updated_at"`
}

type InstanceNode struct {
	ID          int64  `json:"id" db:"id"`
	InstanceID  int64  `json:"instance_id" db:"instance_id"`
	NodeID      int64  `json:"node_id" db:"node_id"`
	Status      string `json:"status" db:"status"`
	ActivatedAt Time   `json:"activated_at,omitempty" db:"activated_at"`
	CompletedAt Time   `json:"completed_at,omitempty" db:"completed_at"`
}

type Task struct {
	ID             int64  `json:"id" db:"id"`
	InstanceID     int64  `json:"instance_id" db:"instance_id"`
	InstanceNodeID int64  `json:"instance_node_id" db:"instance_node_id"`
	ApproverID     int64  `json:"approver_id" db:"approver_id"`
	Status         string `json:"status" db:"status"`
	Comment        string `json:"comment" db:"comment"`
	ActedAt        Time   `json:"acted_at,omitempty" db:"acted_at"`
	CreatedAt      Time   `json:"created_at" db:"created_at"`

	// 附加展示字段（非数据库列）
	NodeName      string `json:"node_name,omitempty" db:"-"`
	InstanceTitle string `json:"instance_title,omitempty" db:"-"`
}

type Log struct {
	ID         int64  `json:"id" db:"id"`
	InstanceID int64  `json:"instance_id" db:"instance_id"`
	ActorID    *int64 `json:"actor_id,omitempty" db:"actor_id"`
	Action     string `json:"action" db:"action"`
	Detail     string `json:"detail" db:"detail"`
	CreatedAt  Time   `json:"created_at" db:"created_at"`
}
