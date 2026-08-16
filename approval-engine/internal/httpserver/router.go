package httpserver

import (
	"database/sql"
	"net/http"

	"github.com/yorkyu/approval-engine/internal/auth"
	"github.com/yorkyu/approval-engine/internal/service"
)

// NewRouter 组装所有路由。使用 Go 1.22+ 标准库 net/http 的模式路由
// （"METHOD /path/{param}" 语法），刻意不引入第三方 Web 框架——
// 这部分逻辑简单到完全不需要框架，标准库已经够用，也更适合看清楚
// "路由 + 中间件"到底是怎么一回事。
func NewRouter(db *sql.DB, signer *auth.Signer) http.Handler {
	authSvc := service.NewAuthService(db, signer)
	templateSvc := service.NewTemplateService(db)
	engine := service.NewEngine(db)
	querySvc := service.NewInstanceQueryService(db)

	authHandler := NewAuthHandler(authSvc)
	templateHandler := NewTemplateHandler(templateSvc)
	instanceHandler := NewInstanceHandler(engine, querySvc)
	taskHandler := NewTaskHandler(engine, querySvc)
	userHandler := NewUserHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})

	// ---- 公开接口 ----
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)

	// ---- 需要登录的接口 ----
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/me", userHandler.Me)
	protected.HandleFunc("GET /api/users", userHandler.List)

	protected.HandleFunc("POST /api/templates", templateHandler.Create)
	protected.HandleFunc("GET /api/templates", templateHandler.List)
	protected.HandleFunc("GET /api/templates/{id}", templateHandler.Get)

	protected.HandleFunc("POST /api/instances", instanceHandler.Start)
	protected.HandleFunc("GET /api/instances/mine", instanceHandler.ListMine)
	protected.HandleFunc("GET /api/instances", instanceHandler.ListAll)
	protected.HandleFunc("GET /api/instances/{id}", instanceHandler.Get)

	protected.HandleFunc("GET /api/tasks/mine", taskHandler.ListMine)
	protected.HandleFunc("POST /api/tasks/{id}/approve", taskHandler.Approve)
	protected.HandleFunc("POST /api/tasks/{id}/reject", taskHandler.Reject)

	mux.Handle("/api/", Chain(protected, RequireAuth(signer)))

	return Chain(mux, WithRecover, WithLogging, WithCORS)
}
