package main

import (
	"context"
	"testing"

	"github.com/ductran999/shared-pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() *todoService {
	return newTodoService(&inMemoryTodoRepository{todos: []Todo{
		{ID: 1, Title: "write demo"},
		{ID: 2, Title: "commit", Done: true},
	}})
}

func Test_TodoService_Get(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		todo, err := service.Get(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, todo.ID)
		assert.Equal(t, "write demo", todo.Title)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := service.Get(ctx, 999)
		require.Error(t, err)

		appErr, ok := errs.From(err)
		require.True(t, ok)
		assert.Equal(t, errs.KindNotFound, appErr.Kind)
		assert.Equal(t, errs.Code("TODO_NOT_FOUND"), appErr.Code)
	})
}

func Test_TodoService_Create(t *testing.T) {
	service := newTestService()
	ctx := context.Background()

	t.Run("created", func(t *testing.T) {
		todo, err := service.Create(ctx, Todo{Title: "new task"})
		require.NoError(t, err)
		assert.Equal(t, 3, todo.ID)
		assert.Equal(t, "new task", todo.Title)
	})

	t.Run("missing title", func(t *testing.T) {
		_, err := service.Create(ctx, Todo{})
		require.Error(t, err)

		appErr, ok := errs.From(err)
		require.True(t, ok)
		assert.Equal(t, errs.KindValidation, appErr.Kind)
		assert.Equal(t, errs.Code("TITLE_REQUIRED"), appErr.Code)

		fieldErrors := appErr.Details
		require.Len(t, fieldErrors, 1)
		assert.Equal(t, "title", fieldErrors[0].Field)
	})
}

func Test_TodoService_List(t *testing.T) {
	service := newTestService()

	items := service.List(context.Background())
	assert.Len(t, items, 2)
}
