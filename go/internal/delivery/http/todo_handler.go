package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/rezafahlevi/gotodo/internal/domain"
)

// TodoHandler handles HTTP requests for Todos.
type TodoHandler struct {
	useCase domain.TodoUseCase
}

// NewTodoHandler creates a new handler and registers the routes.
// We are using the standard library net/http which got much better routing in Go 1.22.
func NewTodoHandler(mux *http.ServeMux, useCase domain.TodoUseCase) {
	handler := &TodoHandler{
		useCase: useCase,
	}

	// Go 1.22 allows HTTP method and path variables in the pattern!
	mux.HandleFunc("GET /todos", handler.FetchAll)
	mux.HandleFunc("GET /todos/facets", handler.GetFacets)
	mux.HandleFunc("GET /todos/{id}", handler.GetByID)
	mux.HandleFunc("POST /todos", handler.Create)
	mux.HandleFunc("PUT /todos/{id}", handler.Update)
	mux.HandleFunc("DELETE /todos/{id}", handler.Delete)
}

func (h *TodoHandler) FetchAll(w http.ResponseWriter, r *http.Request) {
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

	todos, total, err := h.useCase.FetchAll(ctx, searchQuery, statusFilter, page, limit)
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

func (h *TodoHandler) GetFacets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	searchQuery := r.URL.Query().Get("search")
	facets, err := h.useCase.GetFacets(ctx, searchQuery)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, facets)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// r.PathValue extracts variables defined in the mux route pattern (Go 1.22+)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	ctx := r.Context()
	todo, err := h.useCase.GetByID(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Todo not found")
		return
	}
	respondWithJSON(w, http.StatusOK, todo)
}

func validatePasskey(r *http.Request) bool {
	expected := os.Getenv("APP_PASSKEY")
	if expected == "" {
		return true // disabled if not set
	}
	return r.Header.Get("X-Passkey") == expected
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !validatePasskey(r) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid Passkey")
		return
	}

	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()
	todo, err := h.useCase.Create(ctx, input.Title)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !validatePasskey(r) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid Passkey")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var input struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()
	todo, err := h.useCase.Update(ctx, id, input.Title, input.Status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !validatePasskey(r) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid Passkey")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	ctx := r.Context()
	err = h.useCase.Delete(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// Helpers for JSON responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
