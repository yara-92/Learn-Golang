package httpserver

import (
	"database/sql"
	"net/http"

	"github.com/yorkyu/approval-engine/internal/repository"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

// Me 返回当前登录用户的基本信息，前端用来做"我是谁"的展示。
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	u, err := repository.GetUserByID(r.Context(), h.db, claims.UserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	u.PasswordHash = ""
	writeOK(w, u)
}

// List 返回全部用户，主要用于演示/测试时查看有哪些账号可以登录、
// 以及在创建模板时填哪个 user_id 作为具名审批人。
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := repository.ListAllUsers(r.Context(), h.db)
	if err != nil {
		writeErr(w, err)
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	writeOK(w, users)
}
