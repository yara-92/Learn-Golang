package service

import (
	"context"
	"database/sql"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
)

// InstanceQueryService 提供只读查询，不涉及流程状态流转（状态流转都在 Engine 里，
// 因为那部分需要事务 + 互斥锁保证一致性；纯查询没有这个必要）。
type InstanceQueryService struct {
	db *sql.DB
}

func NewInstanceQueryService(db *sql.DB) *InstanceQueryService {
	return &InstanceQueryService{db: db}
}

func (s *InstanceQueryService) Get(ctx context.Context, id int64) (*model.Instance, error) {
	return repository.GetInstance(ctx, s.db, id)
}

func (s *InstanceQueryService) ListMine(ctx context.Context, initiatorID int64) ([]model.Instance, error) {
	return repository.ListInstancesByInitiator(ctx, s.db, initiatorID)
}

func (s *InstanceQueryService) ListAll(ctx context.Context) ([]model.Instance, error) {
	return repository.ListAllInstances(ctx, s.db)
}

func (s *InstanceQueryService) Nodes(ctx context.Context, instanceID int64) ([]model.InstanceNode, error) {
	return repository.ListInstanceNodesByInstance(ctx, s.db, instanceID)
}

func (s *InstanceQueryService) Tasks(ctx context.Context, instanceID int64) ([]model.Task, error) {
	return repository.ListTasksByInstance(ctx, s.db, instanceID)
}

func (s *InstanceQueryService) Logs(ctx context.Context, instanceID int64) ([]model.Log, error) {
	return repository.ListLogsByInstance(ctx, s.db, instanceID)
}

func (s *InstanceQueryService) MyTasks(ctx context.Context, approverID int64, status string) ([]model.Task, error) {
	return repository.ListTasksByApprover(ctx, s.db, approverID, status)
}
