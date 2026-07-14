package domain

import "context"

// Status constants
const (
	StatusOpen     = 1
	StatusProgress = 2
	StatusReview   = 3
	StatusDone     = 4
	StatusHold     = 5
	StatusCanceled = 6
)

// Todo represents a single task in our application.
// We use JSON tags to control how this struct is serialized to the frontend.
type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status int    `json:"status"`
}

// Facet represents the count of tasks for a specific status.
type Facet map[int]int

// TodoRepository defines the interface for data access.
// Clean Architecture rule: The domain layer defines interfaces,
// and outer layers (like repository) implement them.
// We use context.Context for timeout and cancellation control.
type TodoRepository interface {
	FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]Todo, int, error)
	GetFacets(ctx context.Context, searchQuery string) (Facet, error)
	GetByID(ctx context.Context, id int) (Todo, error)
	Create(ctx context.Context, todo *Todo) error
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id int) error
}

// TodoUseCase defines the business logic interface.
// The delivery layer (HTTP handlers) will call these methods.
type TodoUseCase interface {
	FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]Todo, int, error)
	GetFacets(ctx context.Context, searchQuery string) (Facet, error)
	GetByID(ctx context.Context, id int) (Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Update(ctx context.Context, id int, title string, status int) (Todo, error)
	Delete(ctx context.Context, id int) error
}
