package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
	"github.com/yorkyu/approval-engine/internal/seed"
	"github.com/yorkyu/approval-engine/internal/service"
	"github.com/yorkyu/approval-engine/internal/store"
)

func TestFanOutFanIn_FullApprovalFlow(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := seed.Run(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	engine := service.NewEngine(db)

	alice, err := repository.GetUserByUsername(ctx, db, "alice")
	if err != nil {
		t.Fatalf("get alice: %v", err)
	}
	bob, err := repository.GetUserByUsername(ctx, db, "bob")
	if err != nil {
		t.Fatalf("get bob: %v", err)
	}
	carol, err := repository.GetUserByUsername(ctx, db, "carol")
	if err != nil {
		t.Fatalf("get carol: %v", err)
	}
	dave, err := repository.GetUserByUsername(ctx, db, "dave")
	if err != nil {
		t.Fatalf("get dave: %v", err)
	}

	templates, err := repository.ListTemplates(ctx, db)
	if err != nil || len(templates) == 0 {
		t.Fatalf("list templates: %v (n=%d)", err, len(templates))
	}
	templateID := templates[0].ID

	instance, err := engine.StartInstance(ctx, service.StartInstanceInput{
		TemplateID:   templateID,
		BusinessType: "EXPENSE",
		BusinessID:   "E-TEST-1",
		Title:        "单元测试报销",
		FormData:     map[string]any{"amount": 100},
		InitiatorID:  alice.ID,
	})
	if err != nil {
		t.Fatalf("start instance: %v", err)
	}
	if instance.Status != model.InstanceStatusRunning {
		t.Fatalf("expected RUNNING right after start, got %s", instance.Status)
	}

	// 经理（bob）通过后，HR 和财务应该同时出现待办任务（fan-out）。
	bobTasks, err := repository.ListTasksByApprover(ctx, db, bob.ID, model.TaskStatusPending)
	if err != nil || len(bobTasks) != 1 {
		t.Fatalf("expected exactly 1 pending task for bob, got %d (err=%v)", len(bobTasks), err)
	}
	if err := engine.ApproveTask(ctx, bobTasks[0].ID, bob.ID, "ok"); err != nil {
		t.Fatalf("bob approve: %v", err)
	}

	carolTasks, err := repository.ListTasksByApprover(ctx, db, carol.ID, model.TaskStatusPending)
	if err != nil || len(carolTasks) != 1 {
		t.Fatalf("expected exactly 1 pending task for carol after bob approved, got %d (err=%v)", len(carolTasks), err)
	}
	daveTasks, err := repository.ListTasksByApprover(ctx, db, dave.ID, model.TaskStatusPending)
	if err != nil || len(daveTasks) != 1 {
		t.Fatalf("expected exactly 1 pending task for dave after bob approved, got %d (err=%v)", len(daveTasks), err)
	}

	// carol 通过，但 dave 还没动作 -> 流程必须仍是 RUNNING（fan-in 未汇合）。
	if err := engine.ApproveTask(ctx, carolTasks[0].ID, carol.ID, "hr ok"); err != nil {
		t.Fatalf("carol approve: %v", err)
	}
	mid, err := repository.GetInstance(ctx, db, instance.ID)
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if mid.Status != model.InstanceStatusRunning {
		t.Fatalf("expected still RUNNING after only carol approved, got %s", mid.Status)
	}

	// dave 也通过 -> 两个分支都完成，流程应自动 APPROVED。
	if err := engine.ApproveTask(ctx, daveTasks[0].ID, dave.ID, "finance ok"); err != nil {
		t.Fatalf("dave approve: %v", err)
	}
	final, err := repository.GetInstance(ctx, db, instance.ID)
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if final.Status != model.InstanceStatusApproved {
		t.Fatalf("expected APPROVED after both branches done, got %s", final.Status)
	}
}

func TestReject_TerminatesWholeInstance(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := seed.Run(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	engine := service.NewEngine(db)
	alice, _ := repository.GetUserByUsername(ctx, db, "alice")
	bob, _ := repository.GetUserByUsername(ctx, db, "bob")
	templates, _ := repository.ListTemplates(ctx, db)

	instance, err := engine.StartInstance(ctx, service.StartInstanceInput{
		TemplateID:   templates[0].ID,
		BusinessType: "EXPENSE",
		BusinessID:   "E-TEST-2",
		Title:        "会被拒绝的报销",
		FormData:     map[string]any{"amount": 999999},
		InitiatorID:  alice.ID,
	})
	if err != nil {
		t.Fatalf("start instance: %v", err)
	}

	bobTasks, err := repository.ListTasksByApprover(ctx, db, bob.ID, model.TaskStatusPending)
	if err != nil || len(bobTasks) != 1 {
		t.Fatalf("expected 1 pending task for bob, got %d (err=%v)", len(bobTasks), err)
	}
	if err := engine.RejectTask(ctx, bobTasks[0].ID, bob.ID, "金额异常，驳回"); err != nil {
		t.Fatalf("reject: %v", err)
	}

	final, err := repository.GetInstance(ctx, db, instance.ID)
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if final.Status != model.InstanceStatusRejected {
		t.Fatalf("expected REJECTED, got %s", final.Status)
	}
}

func TestApproveTask_WrongApprover_Forbidden(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := seed.Run(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	engine := service.NewEngine(db)
	alice, _ := repository.GetUserByUsername(ctx, db, "alice")
	bob, _ := repository.GetUserByUsername(ctx, db, "bob")
	templates, _ := repository.ListTemplates(ctx, db)

	_, err = engine.StartInstance(ctx, service.StartInstanceInput{
		TemplateID: templates[0].ID, BusinessType: "EXPENSE", BusinessID: "E-TEST-3",
		Title: "越权测试", FormData: map[string]any{}, InitiatorID: alice.ID,
	})
	if err != nil {
		t.Fatalf("start instance: %v", err)
	}
	bobTasks, _ := repository.ListTasksByApprover(ctx, db, bob.ID, model.TaskStatusPending)

	// alice（发起人本人）尝试冒充审批 bob 的任务，应该被拒绝。
	err = engine.ApproveTask(ctx, bobTasks[0].ID, alice.ID, "我不是审批人")
	if err != model.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
