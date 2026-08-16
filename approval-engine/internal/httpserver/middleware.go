package httpserver

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yorkyu/approval-engine/internal/auth"
)

type ctxKey int

const claimsCtxKey ctxKey = 1

// Middleware 是标准的 func(http.Handler) http.Handler 形式，可以自由链式组合。
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// WithLogging 记录每个请求的方法、路径、状态码、耗时——生产环境里通常会换成
// 结构化日志（zap/zerolog）+ 采样，这里为了零依赖用标准库 log。
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// WithRecover 兜底捕获 handler 里的 panic，避免一个请求的异常打垮整个进程。
func WithRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// WithCORS 允许本地前端（如 Vite 开发服务器）直接跨域调用，方便你把这个后端
// 接到已有的 Vue 项目上做联调。生产环境应把 "*" 换成明确的前端域名白名单。
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth 校验 Authorization: Bearer <token>，并把解析出的 Claims 放进
// request context，供下游 handler 通过 ClaimsFromContext 取用。
func RequireAuth(signer *auth.Signer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, errorBody{Error: "missing bearer token"})
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := signer.Parse(token)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, errorBody{Error: "invalid or expired token"})
				return
			}
			ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsCtxKey).(*auth.Claims)
	return claims
}
