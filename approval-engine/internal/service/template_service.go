package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
)

type TemplateService struct {
	db *sql.DB
}

func NewTemplateService(db *sql.DB) *TemplateService {
	return &TemplateService{db: db}
}

func (s *TemplateService) Create(ctx context.Context, def *model.TemplateDef) (int64, error) {
	if err := normalizeTemplateDef(def); err != nil {
		return 0, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	id, err := repository.CreateTemplate(ctx, tx, def)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// normalizeTemplateDef 补全默认值并做基本合法性校验，避免明显错误的模板
// （比如没有 START/END 节点、边引用了不存在的 code）落到数据库里才发现。
func normalizeTemplateDef(def *model.TemplateDef) error {
	if def.Name == "" {
		return fmt.Errorf("%w: name is required", model.ErrTemplateInvalid)
	}
	codes := map[string]bool{}
	hasStart, hasEnd := false, false
	for i := range def.Nodes {
		n := &def.Nodes[i]
		if n.Code == "" || n.Name == "" {
			return fmt.Errorf("%w: node code/name is required", model.ErrTemplateInvalid)
		}
		if codes[n.Code] {
			return fmt.Errorf("%w: duplicate node code %q", model.ErrTemplateInvalid, n.Code)
		}
		codes[n.Code] = true

		if n.ApproveType == "" {
			n.ApproveType = model.ApproveTypeAny
		}
		if n.JoinType == "" {
			n.JoinType = model.JoinTypeAny
		}
		switch n.NodeType {
		case model.NodeTypeStart:
			hasStart = true
		case model.NodeTypeEnd:
			hasEnd = true
		case model.NodeTypeApproval:
			if len(n.Approvers) == 0 {
				return fmt.Errorf("%w: approval node %q needs at least one approver", model.ErrTemplateInvalid, n.Code)
			}
		default:
			return fmt.Errorf("%w: unknown node_type %q", model.ErrTemplateInvalid, n.NodeType)
		}
	}
	if !hasStart || !hasEnd {
		return fmt.Errorf("%w: template must contain exactly one START and one END node", model.ErrTemplateInvalid)
	}
	for _, e := range def.Edges {
		if !codes[e.FromCode] || !codes[e.ToCode] {
			return fmt.Errorf("%w: edge references unknown node code (%s -> %s)", model.ErrTemplateInvalid, e.FromCode, e.ToCode)
		}
	}
	return nil
}

func (s *TemplateService) Get(ctx context.Context, id int64) (*model.Template, error) {
	return repository.GetTemplate(ctx, s.db, id)
}

func (s *TemplateService) List(ctx context.Context) ([]model.Template, error) {
	return repository.ListTemplates(ctx, s.db)
}
