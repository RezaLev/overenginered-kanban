package domain

import "context"

// CQRS Query Models (Reads)
type TodoQueryRepository interface {
	FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]Todo, int, error)
	GetFacets(ctx context.Context, searchQuery string) (Facet, error)
	GetByID(ctx context.Context, id int) (Todo, error)
}

type TodoQueryUseCase interface {
	FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]Todo, int, error)
	GetFacets(ctx context.Context, searchQuery string) (Facet, error)
	GetByID(ctx context.Context, id int) (Todo, error)
}


