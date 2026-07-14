package postgres

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezafahlevi/gotodo/internal/domain"
)

type todoQueryRepository struct {
	db *pgxpool.Pool
}

func NewTodoQueryRepository(db *pgxpool.Pool) domain.TodoQueryRepository {
	return &todoQueryRepository{db: db}
}

func (r *todoQueryRepository) FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]domain.Todo, int, error) {
	offset := (page - 1) * limit
	var total int

	var countQuery string
	var selectQuery string
	var args []interface{}
	argCount := 1

	if searchQuery != "" {
		countQuery = "SELECT count(*) FROM todos WHERE title ILIKE $" + strconv.Itoa(argCount)
		selectQuery = "SELECT id, title, status FROM todos WHERE title ILIKE $" + strconv.Itoa(argCount)
		args = append(args, "%"+searchQuery+"%")
		argCount++
	} else {
		countQuery = "SELECT count(*) FROM todos WHERE 1=1"
		selectQuery = "SELECT id, title, status FROM todos WHERE 1=1"
	}

	if statusFilter != nil {
		countQuery += " AND status = $" + strconv.Itoa(argCount)
		selectQuery += " AND status = $" + strconv.Itoa(argCount)
		args = append(args, *statusFilter)
		argCount++
	}

	if searchQuery == "" && statusFilter == nil {
		err := r.db.QueryRow(ctx, "SELECT reltuples::bigint FROM pg_class WHERE relname = 'todos'").Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	selectQuery += " ORDER BY status ASC, id DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	todos := []domain.Todo{}
	for rows.Next() {
		var t domain.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Status); err != nil {
			return nil, 0, err
		}
		todos = append(todos, t)
	}
	return todos, total, nil
}

func (r *todoQueryRepository) GetFacets(ctx context.Context, searchQuery string) (domain.Facet, error) {
	var query string
	var args []interface{}

	if searchQuery != "" {
		query = "SELECT status, count(*) FROM todos WHERE title ILIKE $1 GROUP BY status"
		args = append(args, "%"+searchQuery+"%")
	} else {
		query = "SELECT status, count(*) FROM todos GROUP BY status"
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facets := make(domain.Facet)
	for rows.Next() {
		var status, count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		facets[status] = count
	}
	return facets, nil
}

func (r *todoQueryRepository) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	var t domain.Todo
	err := r.db.QueryRow(ctx, "SELECT id, title, status FROM todos WHERE id = $1", id).Scan(&t.ID, &t.Title, &t.Status)
	if err != nil {
		return domain.Todo{}, err
	}
	return t, nil
}
