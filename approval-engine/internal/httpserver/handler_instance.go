package httpserver

import (
	"net/http"
	"strconv"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/service"
)

type InstanceHandler struct {
	engine *service.Engine
	query  *service.InstanceQueryService
}

func NewInstanceHandler(engine *service.Engine, query *service.InstanceQueryService) *InstanceHandler {
	return &InstanceHandler{engine: engine, query: query}
}

type startInstanceRequest struct {
	TemplateID   int64          `json:"template_id"`
	BusinessType string         `json:"business_type"`
	BusinessID   string         `json:"business_id"`
	Title        string         `json:"title"`
	FormData     map[string]any `json:"form_data"`
}

func (h *InstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	var req startInstanceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid request body: " + err.Error()})
		return
	}
	instance, err := h.engine.StartInstance(r.Context(), service.StartInstanceInput{
		TemplateID:   req.TemplateID,
		BusinessType: req.BusinessType,
		BusinessID:   req.BusinessID,
		Title:        req.Title,
		FormData:     req.FormData,
		InitiatorID:  claims.UserID,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeCreated(w, instance)
}

func (h *InstanceHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	list, err := h.query.ListMine(r.Context(), claims.UserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, list)
}

func (h *InstanceHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.query.ListAll(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, list)
}

type instanceDetail struct {
	Instance *model.Instance      `json:"instance"`
	Nodes    []model.InstanceNode `json:"nodes"`
	Tasks    []model.Task         `json:"tasks"`
	Logs     []model.Log          `json:"logs"`
}

func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid id"})
		return
	}
	instance, err := h.query.Get(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	nodes, err := h.query.Nodes(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	tasks, err := h.query.Tasks(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	logs, err := h.query.Logs(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, instanceDetail{Instance: instance, Nodes: nodes, Tasks: tasks, Logs: logs})
}
