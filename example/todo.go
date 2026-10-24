package main

import (
	"context"

	"github.com/ductran999/shared-pkg/errs"
)

// Todo is the demo resource.
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// todoRepository is the domain facing port of the storage layer.
type todoRepository interface {
	All(ctx context.Context) []Todo
	FindByID(ctx context.Context, id int) (Todo, bool)
	Save(ctx context.Context, todo Todo) Todo
}

// todoService is the use case layer. It owns the business rules and returns
// classified errors, it knows nothing about HTTP.
type todoService struct {
	repo todoRepository
}

// newTodoService wires the use case with its dependency.
func newTodoService(repo todoRepository) *todoService {
	return &todoService{repo: repo}
}

// List returns every todo.
func (s *todoService) List(ctx context.Context) []Todo {
	return s.repo.All(ctx)
}

// Get returns a todo by id.
func (s *todoService) Get(ctx context.Context, id int) (Todo, error) {
	todo, found := s.repo.FindByID(ctx, id)
	if !found {
		return Todo{}, errs.NotFound("TODO_NOT_FOUND", "todo not found")
	}

	return todo, nil
}

// Create validates and stores a new todo.
func (s *todoService) Create(ctx context.Context, todo Todo) (Todo, error) {
	if todo.Title == "" {
		return Todo{}, errs.Validation("TITLE_REQUIRED", "title is required",
			errs.FieldError{Field: "title", Message: "is required"},
		)
	}

	return s.repo.Save(ctx, todo), nil
}

// inMemoryTodoRepository is a simple in-memory implementation of the port.
type inMemoryTodoRepository struct {
	todos []Todo
}

// All returns every stored todo.
func (r *inMemoryTodoRepository) All(_ context.Context) []Todo {
	return r.todos
}

// FindByID looks a todo up by id.
func (r *inMemoryTodoRepository) FindByID(_ context.Context, id int) (Todo, bool) {
	for _, todo := range r.todos {
		if todo.ID == id {
			return todo, true
		}
	}

	return Todo{}, false
}

// Save appends a todo with the next id.
func (r *inMemoryTodoRepository) Save(_ context.Context, todo Todo) Todo {
	todo.ID = len(r.todos) + 1
	r.todos = append(r.todos, todo)

	return todo
}
