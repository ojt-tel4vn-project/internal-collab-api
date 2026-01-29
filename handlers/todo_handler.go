package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type TodoHandler struct {
	service services.TodoService
}

func NewTodoHandler(service services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req dtos.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid request", err.Error()))
		return
	}

	todo, err := h.service.CreateTodo(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse("Failed to create todo", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse("Todo created successfully", todo))
}

func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	todos, err := h.service.GetAllTodos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse("Failed to get todos", err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse("Todos retrieved successfully", todos))
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid ID", err.Error()))
		return
	}

	todo, err := h.service.GetTodoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse("Todo not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse("Todo retrieved successfully", todo))
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid ID", err.Error()))
		return
	}

	var req dtos.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid request", err.Error()))
		return
	}

	todo, err := h.service.UpdateTodo(uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse("Failed to update todo", err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse("Todo updated successfully", todo))
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse("Invalid ID", err.Error()))
		return
	}

	if err := h.service.DeleteTodo(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse("Failed to delete todo", err.Error()))
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse("Todo deleted successfully", nil))
}
