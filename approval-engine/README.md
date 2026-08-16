# approval-engine —— Go 版 DAG 审批流引擎

用 Go 重新实现的审批流后端：模板化的 DAG（有向无环图）节点定义、支持并行分支
（fan-out）与汇合（fan-in）、事务 + 锁保证并发安全。零外部框架依赖（仅用
`net/http` 标准库），数据库是内嵌的纯 Go SQLite（无需安装/启动任何数据库
服务），**下载后两条命令即可跑起来**。

## 目录结构

```
approval-engine/
├── cmd/server/main.go          # 程序入口
├── internal/
│   ├── config/                 # 环境变量配置
│   ├── store/                  # SQLite 初始化 + 建表
│   ├── model/                  # 领域模型（含自定义 Time 类型、错误定义）
│   ├── repository/             # 数据访问层（纯 SQL，无业务逻辑）
│   ├── service/
│   │   ├── engine.go           # ★ 核心：DAG 审批引擎（发起/审批/拒绝/推进）
│   │   ├── auth_service.go     # 登录/注册
│   │   ├── template_service.go # 模板创建与校验
│   │   └── instance_service.go # 只读查询
│   ├── auth/                   # bcrypt 密码哈希 + 自实现签名 Token
│   ├── httpserver/              # 路由、中间件、handler
│   └── seed/                   # 首次启动写入演示账号 + 示例模板
├── scripts/demo.sh             # 端到端演示脚本（curl + jq）
└── data/                       # SQLite 数据库文件存放目录
```

## 快速开始

**前置要求**：本机安装 Go 1.22 及以上版本（[go.dev/dl](https://go.dev/dl/)）。
除此之外不需要装任何数据库、不需要 Docker，`go mod tidy` 会在联网状态下
自动拉取两个依赖（`modernc.org/sqlite` 纯 Go SQLite 驱动、`golang.org/x/crypto`
的 bcrypt），首次拉取需要联网几秒钟。

```bash
# 1. 解压后进入目录
cd approval-engine

# 2. 拉取依赖（生成/校正 go.sum）
go mod tidy

# 3. 启动（首次启动会自动建表 + 写入演示数据）
go run ./cmd/server
```

看到如下日志即代表启动成功：

```
[seed] 首次启动，写入演示账号与示例模板 ...
[seed] 完成。演示账号（用户名/密码）：
       admin/admin123  bob(经理)/bob123  carol(HR)/carol123  dave(财务)/dave123  alice(员工)/alice123
approval-engine listening on :8080 (db: ./data/approval.db)
```

也可以用 `make run` / `make build`（见 Makefile）。

## 跑一遍完整演示

另开一个终端（需要 `curl` 和 `jq`）：

```bash
bash scripts/demo.sh
```

这个脚本会完整演示一次"报销审批"的 DAG 流转：

```
start → manager_review（经理，ANY）
             │
     ┌───────┴────────┐
 hr_review          finance_review        ← 并行分支（fan-out）
     └───────┬────────┘
            end (join_type=ALL)           ← 必须两个分支都通过才结束（fan-in）
```

即：alice 发起 → bob（经理）通过后，HR 和财务两个分支**同时**出现待办 →
carol（HR）先通过，流程仍是 RUNNING（财务分支没走）→ dave（财务）也通过后，
两个分支都完成，流程自动置为 `APPROVED`。这正是"DAG 审批引擎"和"线性审批链"
的本质区别。

## 手动用 curl 试

```bash
# 登录拿 token
TOKEN=$(curl -s -X POST localhost:8080/api/auth/login \
  -d '{"username":"alice","password":"alice123"}' | jq -r .token)

# 查看模板
curl -s localhost:8080/api/templates -H "Authorization: Bearer $TOKEN" | jq

# 发起一个审批实例
curl -s -X POST localhost:8080/api/instances \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"template_id":1,"business_type":"EXPENSE","business_id":"E1","title":"测试报销","form_data":{"amount":100}}'

# 查看我的待办任务
curl -s "localhost:8080/api/tasks/mine?status=PENDING" -H "Authorization: Bearer $TOKEN" | jq

# 通过某个任务
curl -s -X POST localhost:8080/api/tasks/1/approve \
  -H "Authorization: Bearer $TOKEN" -d '{"comment":"同意"}'
```

## API 一览

| 方法 | 路径 | 说明 | 需要登录 |
|---|---|---|---|
| POST | `/api/auth/login` | 登录，返回 token | 否 |
| POST | `/api/auth/register` | 注册新用户 | 否 |
| GET  | `/api/me` | 当前用户信息 | 是 |
| GET  | `/api/users` | 用户列表 | 是 |
| POST | `/api/templates` | 创建审批模板（DAG 定义） | 是 |
| GET  | `/api/templates` | 模板列表 | 是 |
| GET  | `/api/templates/{id}` | 模板详情（含节点/边） | 是 |
| POST | `/api/instances` | 发起一个审批实例 | 是 |
| GET  | `/api/instances/mine` | 我发起的实例 | 是 |
| GET  | `/api/instances` | 全部实例 | 是 |
| GET  | `/api/instances/{id}` | 实例详情（含节点状态/任务/日志） | 是 |
| GET  | `/api/tasks/mine?status=` | 我的待办/历史任务 | 是 |
| POST | `/api/tasks/{id}/approve` | 通过某个任务 | 是 |
| POST | `/api/tasks/{id}/reject` | 拒绝某个任务（一票否决，终止整个实例） | 是 |

## 自定义一个新模板

模板用 `code` 而不是数据库 ID 来描述节点和边（因为创建时节点还没落库，拿不到
自增 ID），提交给 `POST /api/templates`：

```json
{
  "name": "请假审批",
  "business_type": "LEAVE",
  "nodes": [
    {"code": "start", "name": "开始", "node_type": "START"},
    {
      "code": "manager", "name": "经理审批", "node_type": "APPROVAL",
      "approve_type": "ANY",
      "approvers": [{"approver_type": "ROLE", "approver_ref": "manager"}]
    },
    {"code": "end", "name": "结束", "node_type": "END"}
  ],
  "edges": [
    {"from_code": "start", "to_code": "manager"},
    {"from_code": "manager", "to_code": "end"}
  ]
}
```

- `node_type`: `START` / `APPROVAL` / `END`，一个模板有且仅有一个 START、一个 END。
- `approve_type`（节点内多审批人）：`ANY`（任一人通过即通过，默认）/ `ALL`（需全部通过）。
- `join_type`（多条入边汇合，即 fan-in）：`ANY`（任一分支到达即可激活，默认）/ `ALL`（需所有前驱分支都通过）。
- `approver_type`: `USER`（`approver_ref` 填用户 ID，如 `"3"`）/ `ROLE`（`approver_ref` 填角色名，如 `"manager"`，会自动展开成该角色下所有用户）。

## 核心设计说明（配合之前 Vue+Supabase 版本对照阅读）

- **并发安全**：`internal/service/engine.go` 文件头有详细注释，解释了为什么这里
  用 `sync.Mutex` + SQLite 事务，而不是 Postgres 的 `SELECT ... FOR UPDATE`，
  以及生产环境切换到 Postgres 时具体要改哪几处。
- **DAG 推进逻辑**：`engine.go` 里的 `advance()` 函数是整个引擎的心脏，处理
  fan-out（一个节点通过后可能同时激活多个下游节点）和 fan-in（`join_type=ALL`
  时必须等所有前驱分支都通过才激活下游）。
- **分层架构**：`handler → service → repository`，`repository` 只做 SQL 拼装，
  不包含任何业务判断；业务规则（谁能审批、ANY/ALL 怎么判定、拒绝要不要终止
  整个流程）全部收在 `service` 层，符合"依赖方向由外向内"的原则。

## 迁移到生产环境的建议

1. **换数据库**：把 `internal/store/store.go` 的 `sql.Open("sqlite", ...)`
   换成 `pgx`，`schema.sql` 里的 `AUTOINCREMENT` 换成 `GENERATED ALWAYS AS
   IDENTITY`，`engine.go` 里的 `sync.Mutex` 换成 `SELECT ... FOR UPDATE`
   行锁（多实例部署时进程内锁不会跨进程生效，这一步是必须的）。
2. **换鉴权**：现在是自实现的 HMAC 签名 Token，生产建议换成标准 JWT 库
   （如 `golang-jwt/jwt`）以获得更好的互操作性，或接入公司已有的 SSO/OAuth2。
3. **加可观测性**：接入 Prometheus 指标、结构化日志（zap）、OpenTelemetry
   链路追踪——这些在 demo 版本里为了保持零依赖被省略了。
4. **容器化部署**：见 `Dockerfile`（多阶段构建，最终镜像基于 `scratch`，
   体积在 15MB 左右）。

## 打包为可执行文件

```bash
make build          # 产出 bin/approval-engine，纯静态二进制，可直接拷到服务器运行
./bin/approval-engine
```
