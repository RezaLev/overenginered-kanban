package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	os "github.com/opensearch-project/opensearch-go/v2"
	"github.com/rezafahlevi/gotodo/internal/domain"
)

type todoUseCase struct {
	todoRepo domain.TodoRepository
	osClient *os.Client
}

func NewTodoUseCase(repo domain.TodoRepository, osClient *os.Client) domain.TodoUseCase {
	return &todoUseCase{
		todoRepo: repo,
		osClient: osClient,
	}
}

func (u *todoUseCase) FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]domain.Todo, int, error) {
	return u.todoRepo.FetchAll(ctx, searchQuery, statusFilter, page, limit)
}

func (u *todoUseCase) GetFacets(ctx context.Context, searchQuery string) (domain.Facet, error) {
	return u.todoRepo.GetFacets(ctx, searchQuery)
}

func (u *todoUseCase) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	return u.todoRepo.GetByID(ctx, id)
}

func (u *todoUseCase) Create(ctx context.Context, title string) (domain.Todo, error) {
	if title == "" {
		return domain.Todo{}, errors.New("title cannot be empty")
	}

	todo := domain.Todo{
		Title:  title,
		Status: domain.StatusOpen,
	}

	err := u.todoRepo.Create(ctx, &todo)
	if err != nil {
		return domain.Todo{}, err
	}

	// Dual-Write to OpenSearch
	titleEscaped := strings.ReplaceAll(todo.Title, `"`, `\"`)
	docBody := fmt.Sprintf(`{"id":%d,"title":"%s","status":%d}`, todo.ID, titleEscaped, todo.Status)
	
	res, err := u.osClient.Index(
		"todos",
		strings.NewReader(docBody),
		u.osClient.Index.WithDocumentID(strconv.Itoa(todo.ID)),
		u.osClient.Index.WithContext(ctx),
		u.osClient.Index.WithRefresh("true"), 
	)
	if err == nil && !res.IsError() {
		res.Body.Close()
	} else if err == nil {
		fmt.Printf("OS Index Error: %s\n", res.String())
		res.Body.Close()
	}

	return todo, nil
}

func (u *todoUseCase) Update(ctx context.Context, id int, title string, status int) (domain.Todo, error) {
	if title == "" {
		return domain.Todo{}, errors.New("title cannot be empty")
	}

	todo, err := u.todoRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Todo{}, err
	}

	todo.Title = title
	todo.Status = status

	err = u.todoRepo.Update(ctx, &todo)
	if err != nil {
		return domain.Todo{}, err
	}

	// Dual-Write to OpenSearch
	titleEscaped := strings.ReplaceAll(todo.Title, `"`, `\"`)
	docBody := fmt.Sprintf(`{"id":%d,"title":"%s","status":%d}`, todo.ID, titleEscaped, todo.Status)
	
	res, err := u.osClient.Index(
		"todos",
		strings.NewReader(docBody),
		u.osClient.Index.WithDocumentID(strconv.Itoa(todo.ID)),
		u.osClient.Index.WithContext(ctx),
		u.osClient.Index.WithRefresh("true"),
	)
	if err == nil && !res.IsError() {
		res.Body.Close()
	} else if err == nil {
		fmt.Printf("OS Update Error: %s\n", res.String())
		res.Body.Close()
	}

	return todo, nil
}

func (u *todoUseCase) Delete(ctx context.Context, id int) error {
	err := u.todoRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Dual-Write to OpenSearch
	res, err := u.osClient.Delete(
		"todos",
		strconv.Itoa(id),
		u.osClient.Delete.WithContext(ctx),
		u.osClient.Delete.WithRefresh("true"),
	)
	if err == nil && !res.IsError() {
		res.Body.Close()
	} else if err == nil {
		fmt.Printf("OS Delete Error: %s\n", res.String())
		res.Body.Close()
	}
	
	return nil
}
