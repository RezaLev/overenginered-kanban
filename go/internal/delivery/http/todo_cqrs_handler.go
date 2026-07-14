package http

import (
	"net/http"
	"strconv"

	"github.com/rezafahlevi/gotodo/internal/domain"
)

type TodoCQRSHandler struct {
	queryUseCase domain.TodoQueryUseCase
}

func NewTodoCQRSHandler(mux *http.ServeMux, queryUseCase domain.TodoQueryUseCase) {
	handler := &TodoCQRSHandler{
		queryUseCase: queryUseCase,
	}

	mux.HandleFunc("GET /cqrs/todos", handler.FetchAll)
	mux.HandleFunc("GET /cqrs/todos/facets", handler.GetFacets)
	mux.HandleFunc("GET /cqrs/todos/{id}", handler.GetByID)
}

func (h *TodoCQRSHandler) FetchAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	searchQuery := r.URL.Query().Get("search")

	var statusFilter *int
	if s := r.URL.Query().Get("status"); s != "" {
		if parsedStatus, err := strconv.Atoi(s); err == nil {
			statusFilter = &parsedStatus
		}
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	todos, total, err := h.queryUseCase.FetchAll(ctx, searchQuery, statusFilter, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":  todos,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (h *TodoCQRSHandler) GetFacets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	searchQuery := r.URL.Query().Get("search")
	facets, err := h.queryUseCase.GetFacets(ctx, searchQuery)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, facets)
}

func (h *TodoCQRSHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	ctx := r.Context()
	todo, err := h.queryUseCase.GetByID(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Todo not found")
		return
	}

	respondWithJSON(w, http.StatusOK, todo)
}
