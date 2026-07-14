package usecase

import (
	"context"

	"github.com/rezafahlevi/gotodo/internal/domain"
)

type todoQueryHandler struct {
	queryRepo domain.TodoQueryRepository
}

func NewTodoQueryHandler(queryRepo domain.TodoQueryRepository) domain.TodoQueryUseCase {
	return &todoQueryHandler{
		queryRepo: queryRepo,
	}
}

func (h *todoQueryHandler) FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]domain.Todo, int, error) {
	return h.queryRepo.FetchAll(ctx, searchQuery, statusFilter, page, limit)
}

func (h *todoQueryHandler) GetFacets(ctx context.Context, searchQuery string) (domain.Facet, error) {
	return h.queryRepo.GetFacets(ctx, searchQuery)
}

func (h *todoQueryHandler) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	return h.queryRepo.GetByID(ctx, id)
}
