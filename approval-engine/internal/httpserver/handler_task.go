package httpserver

import (
	"net/http"
	"strconv"

	"github.com/yorkyu/approval-engine/internal/service"
)

type TaskHandler struct {
	engine *service.Engine
	query  *service.InstanceQueryService
}

func NewTaskHandler(engine *service.Engine, query *service.InstanceQueryService) *TaskHandler {
	return &TaskHandler{engine: engine, query: query}
}

// ListMine 支持 ?status=PENDING 过滤，默认返回全部状态。
func (h *TaskHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	status := r.URL.Query().Get("status")
	list, err := h.query.MyTasks(r.Context(), claims.UserID, status)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, list)
}

type taskActionRequest struct {
	Comment string `json:"comment"`
}

func (h *TaskHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.handleAction(w, r, true)
}

func (h *TaskHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.handleAction(w, r, false)
}

func (h *TaskHandler) handleAction(w http.ResponseWriter, r *http.Request, approve bool) {
	claims := ClaimsFromContext(r.Context())
	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid id"})
		return
	}
	var req taskActionRequest
	// 评论是可选的，请求体为空也允许。
	_ = decodeJSON(r, &req)

	if approve {
		err = h.engine.ApproveTask(r.Context(), taskID, claims.UserID, req.Comment)
	} else {
		err = h.engine.RejectTask(r.Context(), taskID, claims.UserID, req.Comment)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, map[string]string{"result": "ok"})
}
