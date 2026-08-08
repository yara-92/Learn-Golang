
参考了目前常见的 Go 学习项目路线：CLI → 文件/数据持久化 → HTTP/API → 数据库 → 并发 → WebSocket → 缓存 → 微服务/分布式等。([UpGrad][1])

下面一套**可以复制给 AI 的项目提示词**。每个项目我都会故意写得比较详细，让 AI 不只是“给我写一个程序”，而是像一个真实的软件开发任务一样，要求你自己设计、实现、测试和迭代。

---

# 一、整体学习路线

我建议按照下面这个顺序：

| 阶段 | 项目                 | 核心能力                      |
| -- | ------------------ | ------------------------- |
| 01 | 命令行记账本             | Go 基础、struct、slice、map、文件 |
| 02 | 本地密码管理器            | 文件存储、JSON、错误处理            |
| 03 | 日志分析工具             | 文件处理、正则、统计、CLI            |
| 04 | 多线程文件搜索器           | goroutine、channel、并发      |
| 05 | HTTP 文件服务器         | HTTP、路由、静态文件              |
| 06 | RESTful Todo API   | Web API、JSON、CRUD         |
| 07 | 博客后端               | PostgreSQL、分层架构、认证        |
| 08 | URL 短链接服务          | Redis、缓存、并发、数据库           |
| 09 | 实时聊天服务器            | WebSocket、goroutine、连接管理  |
| 10 | 图片处理服务             | 上传、队列、Worker Pool         |
| 11 | 简易消息队列             | TCP、并发、持久化                |
| 12 | 任务调度系统             | Scheduler、Worker、分布式思想    |
| 13 | API Gateway        | 反向代理、限流、熔断                |
| 14 | 微型电商后端             | 微服务、事务、库存                 |
| 15 | Mini Cloud Storage | 对象存储、分片、并发                |
| 16 | 最终项目：分布式任务平台       | 综合能力                      |

**不要急着做后面的项目。**
前 1～6 个项目的意义是让你真正熟悉 Go；7～10 开始进入后端工程；11～16 才逐渐进入系统设计。

---

# 01｜CLI 命令行记账本

### 难度：★☆☆☆☆

这是第一阶段，但我不建议只做一个最简单的 Todo。

直接把下面这段给 AI：

你是一名资深 Go 工程师，同时也是我的编程导师。

我正在学习 Go，希望通过一个真实的小型项目掌握 Go 的基础语法和基本工程能力。

请带我从零开发一个「命令行个人记账本」项目。

## 一、项目背景

假设这是一个真正给个人使用的 CLI 工具。

用户可以在终端中记录每天的收入和支出，例如：

* 午餐：35 元
* 地铁：6 元
* 工资：12000 元
* 买书：89 元

程序需要支持添加、查看、删除、修改和统计账单。

项目名称暂定为：

expense-cli

## 二、技术限制

第一版尽可能只使用 Go 标准库。

暂时不要使用：

* Cobra
* Gin
* GORM
* PostgreSQL
* Redis
* 第三方 CLI 框架

我希望通过这个项目理解 Go 本身。

## 三、核心功能

至少实现：

1. 添加账单
2. 查看全部账单
3. 按收入/支出筛选
4. 按分类筛选
5. 删除账单
6. 修改账单
7. 查看总收入
8. 查看总支出
9. 查看当前余额
10. 按月份统计
11. 按分类统计

每条账单至少包含：

* ID
* 类型 income / expense
* 金额
* 分类
* 描述
* 创建时间

例如：

{
"id": 1001,
"type": "expense",
"amount": 35.5,
"category": "food",
"description": "午餐",
"created_at": "2026-08-08T12:30:00+08:00"
}

## 四、命令设计

希望最终可以类似这样使用：

expense add
expense list
expense list --type expense
expense list --category food
expense delete 1001
expense update 1001
expense summary
expense summary --month 2026-08

请你帮我设计合理的 CLI 参数。

## 五、数据存储

第一版本使用 JSON 文件保存数据：

data/expenses.json

程序重新启动以后数据不能丢失。

需要考虑：

* 文件不存在怎么办
* JSON 文件损坏怎么办
* 写入失败怎么办
* 空文件怎么办
* 并发写入暂时如何处理

## 六、工程结构

不要把所有代码写在 main.go。

请设计一个合理但不过度复杂的目录结构，例如：

cmd/
internal/
pkg/
data/

解释每个目录为什么存在。

## 七、教学方式

不要直接一次性给我完整答案。

请把开发过程拆成多个阶段。

每完成一个阶段：

1. 告诉我本阶段学习目标
2. 给我任务
3. 给我必要的 API 提示
4. 让我自己实现
5. 我提交代码后，你再帮我 Review
6. 指出代码中的 Go 风格问题
7. 最后给出改进版本

特别关注：

* struct
* method
* interface
* slice
* map
* error
* pointer
* JSON
* file I/O
* time.Time
* package
* Go module

## 八、质量要求

最终代码应该像一个真实的小工具，而不是课堂 Demo。

需要考虑：

* 输入是否合法
* 金额不能为负数
* ID 如何生成
* 时间如何处理
* 文件写入失败
* 错误信息是否友好
* 代码是否容易测试

最后请给我设计至少 15 个测试场景。

注意：

不要直接给出完整项目代码。

你现在先告诉我：

「第一阶段应该做什么，以及我需要自己完成哪些代码。」

---

# 02｜本地密码管理器

### 难度：★★☆☆☆

这个项目开始让你接触**敏感数据、加密思想、文件安全**。

请作为我的 Go 项目导师，带我开发第二个项目：

「Local Password Manager」

这是一个运行在本地电脑上的命令行密码管理器。

## 项目目标

用户可以在本地安全地保存网站账号信息。

例如：

* GitHub
* Gmail
* AWS
* MySQL
* 公司内部系统

每条记录包含：

* ID
* 网站名称
* URL
* username
* password
* notes
* created_at
* updated_at

## 核心功能

实现：

password add
password list
password get
password update
password delete
password search

例如：

password add --name github
password get 3
password search github

## 安全要求

不要直接把密码明文保存到 JSON。

请带我理解：

* hashing 和 encryption 的区别
* 什么数据需要加密
* master password 的作用
* 密钥派生
* salt
* nonce
* 为什么不能自己设计加密算法

第一版可以使用 Go 标准库提供的密码学能力。

## 数据格式

设计一个合理的数据文件格式。

要求：

* 文件不能直接暴露明文密码
* master password 不应该被保存
* 程序启动时需要验证 master password

## 工程要求

学习：

* interface
* io.Reader
* io.Writer
* error wrapping
* crypto
* JSON
* 文件权限
* environment variable
* os.File

## 异常情况

必须考虑：

* 第一次运行
* 密码错误
* 数据文件不存在
* 数据文件损坏
* 数据文件被修改
* 用户忘记 master password
* 写文件失败
* 权限不足

## 教学方式

不要直接给完整答案。

把项目拆成：

Phase 1：数据模型
Phase 2：CLI
Phase 3：文件存储
Phase 4：密码保护
Phase 5：加密
Phase 6：测试
Phase 7：安全 Review

每个阶段给我任务。

如果我的设计存在明显安全问题，不要替我修复，而是先指出风险，让我思考。

最终目标不是做出一个“看起来能用”的 Demo，而是理解一个本地安全软件为什么需要这样设计。

---

# 03｜日志分析器

### 难度：★★☆☆☆

这个项目非常实用。

以后你会发现，**处理日志是后端工程师非常常见的事情**。

请带我使用 Go 开发一个真实的「服务器日志分析工具」。

项目名称：

log-analyzer

## 背景

假设我管理一台 Web 服务器，每天产生大量访问日志。

例如：

2026-08-08 10:21:31 GET /api/users 200 123ms
2026-08-08 10:21:32 GET /api/users 200 98ms
2026-08-08 10:21:35 POST /api/login 401 32ms

我希望通过 Go 程序分析这些日志。

## 第一阶段

实现：

log-analyzer access.log

输出：

* 总请求数量
* 成功请求数量
* 4xx 数量
* 5xx 数量
* 平均响应时间
* 最大响应时间
* 最慢的 10 个请求
* 请求最多的 URL
* 状态码分布

## 第二阶段

增加：

--from
--to
--status
--method
--path

例如：

log-analyzer access.log --status 500

## 第三阶段

支持 CSV / JSON 输出。

例如：

--format json

## 第四阶段

面对大文件。

假设日志文件：

100MB
1GB
10GB

程序不能一次性把整个文件读取到内存。

要求研究：

bufio.Scanner
bufio.Reader
stream processing

## 第五阶段

增加并发能力。

例如：

一个目录下面有：

logs/
app-01.log
app-02.log
app-03.log
app-04.log

程序并发分析多个日志文件。

要求使用：

goroutine
channel
sync.WaitGroup

## 重点学习

请重点指导我：

* 文件 I/O
* streaming
* parsing
* error handling
* struct
* map
* sorting
* goroutine
* channel
* mutex
* benchmark

最后增加 benchmark：

比较：

单线程分析
多 goroutine 分析

并解释为什么“使用 goroutine 不一定更快”。

---

# 04｜并发文件搜索器

### 难度：★★★☆☆

这里开始真正学习 Go 的招牌能力——**并发**。

请带我开发一个类似 grep / ripgrep 的简化版文件搜索工具：

gosearch

## 使用方式

例如：

gosearch "database" ./project

程序递归搜索目录下所有文本文件。

输出：

文件名
行号
匹配内容

例如：

internal/user/service.go:42
db.QueryContext(ctx, ...)

## 基础功能

实现：

* 递归目录
* 文件过滤
* 文本搜索
* 行号
* 匹配次数
* 大小写敏感/不敏感

## 第二阶段

增加：

--ext .go
--ignore vendor
--ignore node_modules
--case-insensitive
--count

## 第三阶段

实现并发搜索。

要求：

目录扫描器负责发现文件。

多个 worker 负责读取文件。

整体采用：

Producer
Worker Pool
Consumer

设计合理的 channel。

## 必须讨论

* goroutine 数量如何确定
* channel 是否需要 buffer
* worker 数量是否应该等于 CPU 核数
* I/O bound 和 CPU bound
* 如何避免 goroutine 泄漏
* 如何优雅停止搜索
* context.Context 的作用

## 第四阶段

支持：

Ctrl+C 中断。

要求：

使用 context cancellation。

## 第五阶段

增加性能测试：

* 1 worker
* 2 workers
* 4 workers
* 8 workers
* 16 workers

使用 benchmark 或实际测试比较性能。

不要默认认为 worker 越多越快。

最后要求你帮我做一次完整的 concurrency code review。

---

# 05｜HTTP 文件服务器

### 难度：★★★☆☆

这个项目非常适合从 CLI 转向 Web。

请带我使用 Go 标准库开发一个真实的 HTTP 文件服务器。

项目名称：

gofile-server

## 使用场景

假设我有一个目录：

./storage

希望启动：

gofile-server --dir ./storage --port 8080

然后通过浏览器访问：

[http://localhost:8080](http://localhost:8080)

可以浏览目录和下载文件。

## 功能

实现：

GET /
GET /files/*
GET /download/*
HEAD /files/*

支持：

* 文件列表
* 文件下载
* Content-Type
* Content-Length
* Last-Modified
* Range 请求

## 安全要求

必须防止：

../

目录穿越攻击。

例如：

GET /files/../../etc/passwd

不能读取 storage 之外的文件。

## 第二阶段

增加上传：

POST /upload

要求：

* 文件大小限制
* 文件类型检查
* 文件名安全处理
* 临时文件
* 上传失败清理

## 第三阶段

增加：

* request ID
* access log
* recovery middleware
* timeout
* graceful shutdown

## 第四阶段

学习：

net/http
http.Handler
ServeMux
middleware
context.Context
http.Server

尽量使用 Go 标准库。

不要一开始使用 Gin。

最终让我理解：

一个 HTTP 请求从进入服务器，到 handler，再到 response，完整经历了什么。

---

# 06｜RESTful Todo API

### 难度：★★★☆☆

到这里开始进入真正的后端开发。

请带我开发一个生产环境风格的 Todo REST API。

项目名称：

todo-api

## API

POST /api/v1/todos
GET /api/v1/todos
GET /api/v1/todos/{id}
PATCH /api/v1/todos/{id}
DELETE /api/v1/todos/{id}

Todo：

id
title
description
completed
priority
due_date
created_at
updated_at

## 要求

实现：

* JSON request
* JSON response
* 参数校验
* HTTP status code
* 错误响应
* pagination
* filtering
* sorting

例如：

GET /api/v1/todos?page=2&page_size=20
GET /api/v1/todos?completed=false
GET /api/v1/todos?priority=high

## 架构

不要把所有代码放在 handler。

尝试设计：

Handler
Service
Repository

例如：

handler/
service/
repository/
model/

让我理解为什么需要分层。

## 数据库

第二阶段使用 PostgreSQL。

需要设计：

users
todos

表结构。

## 测试

要求编写：

unit test
handler test
repository test
integration test

重点学习：

httptest
database/sql
context.Context
transactions

最后让我自己写一份 API 文档。

---

# 07｜博客系统后端

### 难度：★★★★☆

这个项目开始接近**真正的后端项目**。

请作为资深 Go 后端工程师，带我从零开发一个生产级博客系统后端。

项目名称：

blog-service

## 用户系统

用户可以：

注册
登录
修改个人资料
修改密码

字段：

id
username
email
password_hash
created_at
updated_at

## 文章系统

文章：

id
author_id
title
slug
content
status
created_at
updated_at
published_at

状态：

draft
published

## API

POST /api/v1/auth/register
POST /api/v1/auth/login
GET /api/v1/posts
GET /api/v1/posts/{slug}
POST /api/v1/posts
PATCH /api/v1/posts/{id}
DELETE /api/v1/posts/{id}

## 权限

只有作者本人可以：

修改文章
删除文章
发布文章

管理员可以管理所有文章。

## 技术

Go
PostgreSQL
Redis（后续阶段）

## 必须学习

* password hashing
* authentication
* authorization
* JWT 或 session 的取舍
* middleware
* transaction
* database index
* pagination
* validation
* structured logging

## 架构

要求逐渐演化：

第一版：

handler → service → repository

第二版：

加入：

config
logger
middleware
database
cache

第三版：

加入：

Docker
docker-compose
migration
health check

不要一开始就使用过度复杂的架构。

让我体验项目从“小程序”逐渐成长为“真实后端服务”的过程。

---

# 08｜URL 短链接服务

### 难度：★★★★☆

这是一个非常好的系统设计入门项目。

请带我开发一个类似 TinyURL 的 URL Shortener。

项目名称：

shortener

## 基本功能

用户提交：

[https://example.com/articles/2026/08/08/very-long-url](https://example.com/articles/2026/08/08/very-long-url)

服务器返回：

[https://short.local/a8Kx2P](https://short.local/a8Kx2P)

访问：

GET /a8Kx2P

自动 302/301 跳转到原始 URL。

## 数据模型

设计：

id
short_code
original_url
created_at
expires_at
click_count

## 第一阶段

只使用 PostgreSQL。

## 第二阶段

加入 Redis。

访问短链接时：

Redis hit
↓
直接返回

Redis miss
↓
PostgreSQL
↓
写入 Redis

## 第三阶段

考虑高并发。

假设：

1000 req/s
10000 req/s

讨论：

* 数据库压力
* Redis
* connection pool
* cache strategy
* hot key

## 第四阶段

加入统计：

* 总点击次数
* 每日点击
* User-Agent
* Referer
* IP（讨论隐私问题）

## 第五阶段

设计防滥用：

* URL 校验
* rate limiting
* expiration
* 最大 URL 长度
* 黑名单

重点不是“把功能写出来”。

重点是让我理解：

为什么一个看起来只有几十行逻辑的短链接服务，在高并发以后会变成一个系统设计问题。

---

# 09｜实时聊天服务器

### 难度：★★★★☆

这个项目专门训练 **goroutine + channel + WebSocket**。

请带我开发一个实时聊天系统。

项目：

chat-server

## 功能

用户可以：

注册
登录
进入聊天室
发送消息
接收消息
离开聊天室

## WebSocket

建立：

/ws

客户端连接后保持长连接。

服务器需要维护：

用户
连接
聊天室
在线状态

## 消息格式

例如：

{
"type": "message",
"room_id": "general",
"sender": "alice",
"content": "hello",
"timestamp": "..."
}

## 并发模型

重点让我自己设计：

每个连接是否需要 goroutine？

是否需要：

readPump
writePump

聊天室如何广播？

channel 如何设计？

## 必须解决

* 慢客户端
* 断线
* 心跳
* ping/pong
* 写超时
* 连接关闭
* goroutine 泄漏
* 广播阻塞
* mutex
* channel close

## 第二阶段

消息持久化到 PostgreSQL。

## 第三阶段

增加 Redis Pub/Sub。

目标：

Server A
Server B
Server C

用户连接在不同服务器上也能聊天。

## 最终目标

让我理解：

单机实时系统

如何一步步演变成

多节点实时系统。

---

# 10｜图片处理服务

### 难度：★★★★★

这里开始进入真正的**异步任务系统**。

请带我开发一个图片处理后端服务。

项目：

image-service

## 使用场景

用户上传图片：

POST /api/v1/images

服务器返回：

{
"id": "img_123",
"status": "processing"
}

后台异步执行：

* resize
* thumbnail
* metadata extraction
* format conversion

处理完成：

status = completed

## API

POST /images
GET /images/{id}
GET /images/{id}/status
DELETE /images/{id}

## 核心架构

不要让 HTTP 请求直接完成所有图片处理。

设计：

HTTP Server
↓
Task Queue
↓
Worker Pool
↓
Image Processor
↓
Storage

## Worker Pool

要求学习：

goroutine
channel
worker
job
context
retry

## 失败处理

如果图片处理失败：

不要直接丢失任务。

设计：

pending
processing
completed
failed

失败任务支持 retry。

## 第二阶段

加入：

最大并发 worker 数
任务超时
retry count
exponential backoff

## 第三阶段

加入 Redis 作为任务队列。

## 第四阶段

考虑：

Worker 崩溃怎么办？

任务执行到一半服务器死机怎么办？

任务是否可能执行两次？

如何做到幂等？

请重点从真实生产环境角度指导我，而不是只写 Demo。

---

# 11｜自己实现一个简易消息队列

### 难度：★★★★★

这个项目非常适合深入 Go。

请带我从零设计并实现一个简化版 Message Queue。

项目：

mini-queue

目标不是复刻 Kafka。

而是通过自己实现一个小型消息队列理解：

producer
consumer
topic
partition
offset
ack
retry
persistence

## 基础 API

Producer：

Publish(topic, message)

Consumer：

Subscribe(topic)

## 第一阶段

所有数据只保存在内存。

## 第二阶段

实现磁盘持久化。

例如：

data/topic-001.log

消息追加写入。

## 第三阶段

Consumer 支持 offset。

例如：

offset = 1024

Consumer 重启后可以继续读取。

## 第四阶段

增加：

多个 consumer
consumer group
ack
retry

## 第五阶段

模拟：

Producer A
Producer B

Consumer A
Consumer B

讨论消息如何分配。

## 第六阶段

增加并发。

要求重点学习：

* goroutine
* channel
* mutex
* atomic
* sync.Cond
* context
* file I/O
* binary encoding

## 最终挑战

模拟：

10000 条消息

100 个 producer

20 个 consumer

然后 benchmark。

要求我解释：

吞吐量
延迟
消息顺序
数据可靠性

最后和 Kafka 的设计思想做概念层面对比。

---

# 12｜任务调度系统

### 难度：★★★★★

这个项目已经很接近真实后台基础设施。

请带我开发一个分布式任务调度系统。

项目：

task-scheduler

## 用户可以创建任务

例如：

每天 09:00：

发送日报

每小时：

同步数据

每 10 分钟：

检查服务状态

## API

POST /tasks
GET /tasks
GET /tasks/{id}
DELETE /tasks/{id}

任务：

id
name
schedule
payload
status
last_run_at
next_run_at

## Scheduler

Scheduler 负责：

发现即将执行的任务

然后：

Task Queue
↓
Worker

## Worker

多个 worker 并行执行任务。

## 必须解决

* 一个任务不能重复执行
* Worker 崩溃
* Scheduler 崩溃
* 任务执行超时
* retry
* backoff
* 任务取消
* graceful shutdown

## 分布式版本

运行：

Scheduler A
Scheduler B

Worker 1
Worker 2
Worker 3

要求讨论：

如果 A 和 B 同时发现同一个任务怎么办？

研究：

distributed lock

数据库锁

lease

leader election

## 最终目标

实现一个可以：

创建任务
调度任务
执行任务
重试任务
查看任务状态

的小型分布式任务平台。

---

# 13｜API Gateway

### 难度：★★★★★

这时候你已经可以开始接触真正的基础设施开发。

请带我使用 Go 开发一个简化版 API Gateway。

项目：

go-gateway

## 架构

Client
↓
API Gateway
↓
Service A
Service B
Service C

## 基础功能

Gateway 根据路径转发请求：

/api/users/*
→ user-service

/api/orders/*
→ order-service

/api/products/*
→ product-service

## 功能

实现：

* reverse proxy
* routing
* request logging
* request ID
* timeout
* retry
* rate limiting
* authentication
* health check

## 第二阶段

增加：

load balancing

例如：

user-service:

10.0.0.1
10.0.0.2
10.0.0.3

Gateway 自动选择后端。

实现：

round robin

## 第三阶段

增加：

circuit breaker

如果某个服务连续失败：

Gateway 暂时停止向它发送请求。

## 第四阶段

增加：

metrics

统计：

request count
error count
latency
status code

## 最终

使用 Docker 启动：

gateway
user-service
order-service
product-service

让我理解：

API Gateway 为什么存在，以及它和 Nginx、Load Balancer、Service Mesh 分别解决什么问题。

---

# 14｜微型电商系统

### 难度：★★★★★★

到了这里才真正开始碰**复杂业务系统**。

请带我设计并实现一个真实的电商后端系统。

项目名称：

mini-commerce

## 用户

注册
登录
地址管理

## 商品

商品
SKU
价格
库存

## 购物车

添加商品
删除商品
修改数量

## 订单

创建订单
支付
取消订单
查询订单

订单状态：

pending
paid
shipped
completed
cancelled

## 最重要的问题

库存。

例如：

商品库存只有：

10

同时有：

100 个用户购买。

必须保证：

库存不能变成 -1。

## 需要研究

database transaction

row-level lock

optimistic locking

pessimistic locking

Redis

distributed lock

## 服务拆分

第一版：

单体应用。

第二版：

拆成：

user-service
product-service
order-service
payment-service

## 必须处理

订单创建失败怎么办？

库存扣减成功但订单创建失败怎么办？

支付成功但服务崩溃怎么办？

重复支付怎么办？

用户重复提交订单怎么办？

## 最终目标

不要追求“代码很多”。

我要理解：

真实业务系统为什么难。

尤其是：

一致性
事务
并发
幂等
失败恢复

这些问题。

---

# 15｜Mini Cloud Storage

### 难度：★★★★★★

这个项目非常适合 Go。

请带我开发一个简化版云存储系统。

类似一个非常简化的：

Google Drive / Dropbox / S3

项目：

mini-storage

## 基础功能

用户上传：

POST /files

下载：

GET /files/{id}

删除：

DELETE /files/{id}

查看：

GET /files

## 文件信息

id
name
size
content_type
hash
owner_id
created_at

## 第一阶段

文件存本地磁盘。

## 第二阶段

实现：

文件 hash

如果两个用户上传完全相同的文件：

不要重复保存。

## 第三阶段

大文件上传。

不能：

io.ReadAll()

一次把整个文件读进内存。

要求使用 streaming。

## 第四阶段

Multipart Upload。

把一个大文件：

1GB

拆成：

100MB × 10

用户可以分别上传。

## 第五阶段

断点续传。

如果第 7 个 chunk 上传失败：

只重新上传第 7 个。

## 第六阶段

并发下载。

增加：

Range Request

## 第七阶段

设计存储层 interface：

type Storage interface {
Put(...)
Get(...)
Delete(...)
}

然后实现：

LocalStorage

未来可以扩展：

S3Storage

## 最终目标

让我真正理解：

对象存储为什么适合大规模文件系统。

---

# 16｜终极项目：分布式任务平台

### 难度：★★★★★★★

如果前面 15 个项目你认真完成了，最后可以做这个。

这个项目不要一次做完，而是做 **V1 → V2 → V3 → V4**。

请作为我的资深 Go 后端导师，带我完成一个最终项目：

「Distributed Task Platform」

这是我学习 Go 后的毕业项目。

我希望最终做出一个类似简化版：

Celery
Temporal
Airflow

的任务平台。

但不要直接照搬这些系统。

我要自己理解它。

## 一、核心需求

用户可以创建任务：

例如：

任务：

"generate-report"

参数：

{
"user_id": 123,
"month": "2026-08"
}

平台负责：

创建任务
排队
调度
执行
记录状态
失败重试
查看日志

## 二、任务生命周期

设计：

created
queued
running
success
failed
retrying
cancelled

## 三、系统组件

最终系统包含：

API Server

Scheduler

Message Queue

Worker

Database

Redis

Object Storage

Monitoring

## 四、架构

最终：

Client
↓
API Server
↓
Database
↓
Scheduler
↓
Message Queue
↓
Worker Pool
↓
Task Execution

Worker 执行结果：

成功 → Database

失败 → Retry Queue

日志 → Object Storage

Metrics → Monitoring

## 五、第一阶段

先不要微服务。

全部做成：

Go Monolith

实现：

创建任务
执行任务
查询任务

## 六、第二阶段

拆出 Worker。

API Server：

负责接收任务。

Worker：

负责执行任务。

通过 Queue 通信。

## 七、第三阶段

多个 Worker。

例如：

worker-01
worker-02
worker-03
worker-04

任务可以被任意 worker 执行。

## 八、第四阶段

增加：

retry

exponential backoff

timeout

dead letter queue

## 九、第五阶段

解决：

Worker 执行任务时突然崩溃怎么办？

任务会不会永久丢失？

任务会不会执行两次？

如何实现：

at-least-once delivery？

## 十、第六阶段

增加幂等。

例如：

同一个任务：

task_id = 12345

即使执行两次：

也不能导致业务数据错误。

## 十一、第七阶段

增加监控：

Prometheus metrics

例如：

tasks_total
tasks_success_total
tasks_failed_total
task_duration_seconds
queue_depth
worker_active

## 十二、第八阶段

加入 Docker。

最终使用：

docker compose up

启动：

api
scheduler
worker
postgres
redis

## 十三、第九阶段

进行压力测试。

模拟：

10000 tasks

100 workers

100 concurrent clients

测量：

throughput
latency
failure rate
queue delay

## 十四、最终要求

完成项目以后，请让我写：

1. 系统架构图
2. 数据库 ER 图
3. API 文档
4. Queue 设计说明
5. Worker 设计说明
6. Retry 机制说明
7. Failure Recovery 说明
8. Idempotency 说明
9. 性能测试报告
10. 技术选型说明

最重要：

不要为了让我“做出来”而直接给答案。

把你自己当成 Tech Lead。

我负责写代码。

你负责：

拆任务
提出问题
Review
指出设计缺陷
设计测试
模拟生产故障
要求我解释自己的设计

最终让我具备独立设计 Go 后端系统的能力。

---

# 我特别建议你这样使用这些提示词

不要把这些 Prompt 当成：

> “让 AI 帮我写代码。”

而应该当成：

> **“让 AI 当我的 Go Tech Lead，我负责写代码。”**

这两种学习效果差别非常大。

例如你完成第 4 个项目时，不要问：

> “帮我写一个 goroutine worker pool。”

而应该问：

> “我设计了一个 8 worker 的模型，这是我的代码，请 Review。不要直接修改代码，先告诉我有哪些并发问题。”

这样才会真正训练能力。

---

## 一条非常重要的升级路线

我会建议你把每个项目都进行 **4 次迭代**：

**V1：能运行**

↓

**V2：代码重构**

↓

**V3：加入测试**

↓

**V4：模拟真实生产环境**

比如「Todo API」：

### V1

```text
POST /todos
GET /todos
DELETE /todos
```

### V2

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
PostgreSQL
```

### V3

```text
Unit Test
Integration Test
HTTP Test
Database Test
```

### V4

开始故意制造问题：

```text
数据库挂了怎么办？

Redis 挂了怎么办？

请求重复提交怎么办？

1000 个请求同时进来怎么办？

某个请求执行 30 秒怎么办？

客户端中途断开怎么办？

服务启动时数据库还没启动怎么办？

服务关闭时还有请求怎么办？
```

**这一步非常重要。**

因为真正的 Go 后端能力，不是“会写一个 CRUD”，而是逐渐开始思考**并发、资源、失败、恢复、可观测性和系统边界**。Go 的 goroutine/channel、HTTP、网络编程等能力尤其适合通过这种递进式项目来学习。([Calmops][2])

---

# 如果你是从零开始

我建议你**不要一次做 16 个项目**。

最合理的是：

```text
01 记账本
 ↓
02 密码管理器
 ↓
03 日志分析器
 ↓
04 并发搜索器
 ↓
05 HTTP 文件服务器
 ↓
06 REST API
 ↓
07 Blog
 ↓
08 URL Shortener
 ↓
09 WebSocket Chat
 ↓
10 Image Worker
 ↓
11 Message Queue
 ↓
12 Scheduler
 ↓
13 API Gateway
 ↓
14 E-commerce
 ↓
15 Object Storage
 ↓
16 Distributed Task Platform
```

前 **6 个**让你真正掌握 Go。

**7～10**让你具备 Go 后端开发能力。

**11～13**开始训练系统设计和并发。

**14～16**则开始接近高级后端/基础设施工程师的工作方式。

而且这套顺序不是单纯按“项目听起来有多酷”排序，而是让后一个项目尽量复用前一个项目学到的东西；这种从 CLI、API、数据库一路走向并发、缓存、WebSocket 和分布式系统的渐进路线，也与当前 Go 项目路线中常见的 Beginner → Intermediate → Advanced 划分相吻合。([Golang Mastery][3])


[1]: https://www.upgrad.com/blog/golang-projects-ideas/?utm_source=chatgpt.com "Top 26 Golang Project Ideas & Topics [For Freshers & Experienced]"
[2]: https://calmops.com/programming/golang/?utm_source=chatgpt.com "Go - Calmops | Tech, Business & Indie Hacker Knowledge Base"
[3]: https://golangmastery.github.io/projects/?utm_source=chatgpt.com "Go Programming Projects | GolangMastery"
