package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type TodoHandler struct {
	service services.TodoService
}

func NewTodoHandler(service services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (h *TodoHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-todo",
		Method:      http.MethodPost,
		Path:        "/api/v1/todos",
		Summary:     "Create a new todo",
		Tags:        []string{"Todos"},
	}, h.CreateTodo)

	huma.Register(api, huma.Operation{
		OperationID: "get-all-todos",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos",
		Summary:     "Get all todos",
		Tags:        []string{"Todos"},
	}, h.GetAllTodos)

	huma.Register(api, huma.Operation{
		OperationID: "get-todo-by-id",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Get a todo by ID",
		Tags:        []string{"Todos"},
	}, h.GetTodoByID)

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPut,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Update a todo",
		Tags:        []string{"Todos"},
	}, h.UpdateTodo)

	huma.Register(api, huma.Operation{
		OperationID: "delete-todo",
		Method:      http.MethodDelete,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Delete a todo",
		Tags:        []string{"Todos"},
	}, h.DeleteTodo)
}

func (h *TodoHandler) CreateTodo(ctx context.Context, input *struct {
	Body dtos.CreateTodoRequest
}) (*struct {
	Body dtos.Response
}, error) {
	todo, err := h.service.CreateTodo(input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create todo", err)
	}
	return &struct{ Body dtos.Response }{
		Body: dtos.SuccessResponse("Todo created successfully", todo),
	}, nil
}

func (h *TodoHandler) GetAllTodos(ctx context.Context, input *struct{}) (*struct {
	Body dtos.Response
}, error) {
	todos, err := h.service.GetAllTodos()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get todos", err)
	}
	return &struct{ Body dtos.Response }{
		Body: dtos.SuccessResponse("Todos retrieved successfully", todos),
	}, nil
}

func (h *TodoHandler) GetTodoByID(ctx context.Context, input *struct {
	ID uint `path:"id" doc:"Todo ID"`
}) (*struct {
	Body dtos.Response
}, error) {
	todo, err := h.service.GetTodoByID(input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("Todo not found", err)
	}
	return &struct{ Body dtos.Response }{
		Body: dtos.SuccessResponse("Todo retrieved successfully", todo),
	}, nil
}

func (h *TodoHandler) UpdateTodo(ctx context.Context, input *struct {
	ID   uint `path:"id" doc:"Todo ID"`
	Body dtos.UpdateTodoRequest
}) (*struct {
	Body dtos.Response
}, error) {
	todo, err := h.service.UpdateTodo(input.ID, input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update todo", err)
	}
	return &struct{ Body dtos.Response }{
		Body: dtos.SuccessResponse("Todo updated successfully", todo),
	}, nil
}

func (h *TodoHandler) DeleteTodo(ctx context.Context, input *struct {
	ID uint `path:"id" doc:"Todo ID"`
}) (*struct {
	Body dtos.Response
}, error) {
	if err := h.service.DeleteTodo(input.ID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete todo", err)
	}
	return &struct{ Body dtos.Response }{
		Body: dtos.SuccessResponse("Todo deleted successfully", nil),
	}, nil
}
