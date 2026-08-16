#!/usr/bin/env bash
# 一个端到端演示脚本：alice 发起报销 -> bob(经理)审批 -> carol(HR) + dave(财务)
# 并行审批 -> 流程自动结束。用来验证整套 DAG 引擎是否工作正常。
#
# 用法：先在另一个终端 `make run` 或 `go run ./cmd/server` 启动服务，
# 再执行 `bash scripts/demo.sh`（需要本机安装 curl 和 jq）。
set -euo pipefail

BASE="http://localhost:8080"

need() { command -v "$1" >/dev/null 2>&1 || { echo "需要先安装 $1"; exit 1; }; }
need curl
need jq

login() {
  curl -s -X POST "$BASE/api/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$1\",\"password\":\"$2\"}" | jq -r '.token'
}

echo "== 1. 登录 alice（员工）=="
ALICE_TOKEN=$(login alice alice123)
echo "token: ${ALICE_TOKEN:0:16}..."

echo "== 2. 查询模板列表，取第一个模板 ID =="
TEMPLATE_ID=$(curl -s "$BASE/api/templates" -H "Authorization: Bearer $ALICE_TOKEN" | jq -r '.[0].id')
echo "template_id: $TEMPLATE_ID"

echo "== 3. alice 发起一笔报销申请 =="
INSTANCE_ID=$(curl -s -X POST "$BASE/api/instances" \
  -H "Authorization: Bearer $ALICE_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"template_id\":$TEMPLATE_ID,\"business_type\":\"EXPENSE\",\"business_id\":\"EXP-2026-001\",\"title\":\"8月差旅报销\",\"form_data\":{\"amount\":1280.5,\"reason\":\"客户拜访差旅\"}}" \
  | jq -r '.id')
echo "instance_id: $INSTANCE_ID"

echo "== 4. 登录 bob（经理），查看待办并通过 =="
BOB_TOKEN=$(login bob bob123)
BOB_TASK_ID=$(curl -s "$BASE/api/tasks/mine?status=PENDING" -H "Authorization: Bearer $BOB_TOKEN" | jq -r '.[0].id')
curl -s -X POST "$BASE/api/tasks/$BOB_TASK_ID/approve" \
  -H "Authorization: Bearer $BOB_TOKEN" -H 'Content-Type: application/json' \
  -d '{"comment":"同意，金额合理"}' | jq .

echo "== 5. 查看当前实例状态（此时应处于两个并行分支：HR + 财务）=="
curl -s "$BASE/api/instances/$INSTANCE_ID" -H "Authorization: Bearer $ALICE_TOKEN" \
  | jq '{status: .instance.status, nodes: [.nodes[] | {node_id, status}]}'

echo "== 6. 登录 carol（HR），通过 =="
CAROL_TOKEN=$(login carol carol123)
CAROL_TASK_ID=$(curl -s "$BASE/api/tasks/mine?status=PENDING" -H "Authorization: Bearer $CAROL_TOKEN" | jq -r '.[0].id')
curl -s -X POST "$BASE/api/tasks/$CAROL_TASK_ID/approve" \
  -H "Authorization: Bearer $CAROL_TOKEN" -H 'Content-Type: application/json' \
  -d '{"comment":"HR 无异议"}' | jq .

echo "== 7. 此时流程应仍在 RUNNING（财务分支还没走）=="
curl -s "$BASE/api/instances/$INSTANCE_ID" -H "Authorization: Bearer $ALICE_TOKEN" | jq '.instance.status'

echo "== 8. 登录 dave（财务），通过 -> 两个分支都完成，流程应自动结束 =="
DAVE_TOKEN=$(login dave dave123)
DAVE_TASK_ID=$(curl -s "$BASE/api/tasks/mine?status=PENDING" -H "Authorization: Bearer $DAVE_TOKEN" | jq -r '.[0].id')
curl -s -X POST "$BASE/api/tasks/$DAVE_TASK_ID/approve" \
  -H "Authorization: Bearer $DAVE_TOKEN" -H 'Content-Type: application/json' \
  -d '{"comment":"财务无异议，可以报销"}' | jq .

echo "== 9. 最终状态（期望 APPROVED）=="
curl -s "$BASE/api/instances/$INSTANCE_ID" -H "Authorization: Bearer $ALICE_TOKEN" \
  | jq '{status: .instance.status, logs: [.logs[] | .action]}'
