package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/notification"
	"github.com/ojt-tel4vn-project/internal-collab-api/middleware"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type NotificationHandler struct {
	service    services.NotificationService
	jwtService crypto.JWTService
}

func NewNotificationHandler(service services.NotificationService, jwtService crypto.JWTService) *NotificationHandler {
	return &NotificationHandler{
		service:    service,
		jwtService: jwtService,
	}
}

func (h *NotificationHandler) RegisterRoutes(api huma.API) {
	// List Notifications
	huma.Register(api, huma.Operation{
		OperationID: "notifications-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/notifications",
		Summary:     "Get Notification List",
		Tags:        []string{"Notifications"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetNotifications)

	// Get Unread Count
	huma.Register(api, huma.Operation{
		OperationID: "notifications-unread-count",
		Method:      http.MethodGet,
		Path:        "/api/v1/notifications/unread-count",
		Summary:     "Get Unread Notification Count",
		Tags:        []string{"Notifications"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetUnreadCount)

	// Mark specific notification as read
	huma.Register(api, huma.Operation{
		OperationID: "notifications-mark-read",
		Method:      http.MethodPut,
		Path:        "/api/v1/notifications/{id}/read",
		Summary:     "Mark Notification as Read",
		Tags:        []string{"Notifications"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.MarkAsRead)

	// Notification Preferences
	huma.Register(api, huma.Operation{
		OperationID: "notifications-preferences",
		Method:      http.MethodGet,
		Path:        "/api/v1/notifications/preferences",
		Summary:     "Get Notification Preferences",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.GetPreferences)

	// Notification Types
	huma.Register(api, huma.Operation{
		OperationID: "notifications-types",
		Method:      http.MethodGet,
		Path:        "/api/v1/notifications/types",
		Summary:     "Get Notification Types",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.GetTypes)

	// Mark all as read — POST alias (frontend calls POST)
	huma.Register(api, huma.Operation{
		OperationID: "notifications-mark-all-read-post",
		Method:      http.MethodPost,
		Path:        "/api/v1/notifications/read-all",
		Summary:     "Mark All Notifications as Read (POST)",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.MarkAllAsRead)
}

// Handler Implementation

type GetNotificationsRequest struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer auth token"`
	Page          int    `query:"page,omitempty" default:"1" doc:"Page number"`
	Limit         int    `query:"limit,omitempty" default:"20" doc:"Items per page"`
}

func (h *NotificationHandler) GetNotifications(ctx context.Context, input *GetNotificationsRequest) (*struct {
	Body notification.ListNotificationResponse
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}

	notifications, total, err := h.service.GetNotifications(claims.UserID, input.Page, input.Limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch notifications")
	}

	unreadCount, _ := h.service.GetUnreadCount(claims.UserID)

	notificationList := make([]notification.NotificationResponse, len(notifications))
	for i, n := range notifications {
		notificationList[i] = notification.NotificationResponse{
			ID:         n.ID,
			EmployeeID: n.EmployeeID,
			Type:       n.Type,
			Title:      n.Title,
			Message:    n.Message,
			IsRead:     n.IsRead,
			ReadAt:     n.ReadAt,
			CreatedAt:  n.CreatedAt,
			Priority:   string(n.Priority),
		}
		if n.ActionURL != nil {
			notificationList[i].ActionURL = n.ActionURL
		}
	}

	resp := notification.ListNotificationResponse{
		Notifications: notificationList,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          input.Page,
		Limit:         input.Limit,
	}

	return &struct {
		Body notification.ListNotificationResponse
	}{Body: resp}, nil
}

func (h *NotificationHandler) GetUnreadCount(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*struct {
	Body notification.UnreadCountResponse
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}

	count, err := h.service.GetUnreadCount(claims.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch unread count")
	}

	return &struct {
		Body notification.UnreadCountResponse
	}{Body: notification.UnreadCountResponse{Count: count}}, nil
}

func (h *NotificationHandler) MarkAsRead(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
	ID            string `path:"id" required:"true"`
}) (*struct {
	Body struct {
		Message string `json:"message"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}

	notifID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid ID format")
	}

	if err = h.service.MarkAsRead(notifID, claims.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to mark as read")
	}

	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{Body: struct {
		Message string `json:"message"`
	}{Message: "Marked as read"}}, nil
}

func (h *NotificationHandler) MarkAllAsRead(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*struct {
	Body struct {
		Message string `json:"message"`
	}
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}

	if err = h.service.MarkAllAsRead(claims.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to mark all as read")
	}

	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{Body: struct {
		Message string `json:"message"`
	}{Message: "All marked as read"}}, nil
}

// GetPreferences returns notification preferences (stub - returns empty for now)
func (h *NotificationHandler) GetPreferences(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*struct {
	Body struct {
		Data map[string]bool `json:"data"`
	}
}, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}
	return &struct {
		Body struct {
			Data map[string]bool `json:"data"`
		}
	}{Body: struct {
		Data map[string]bool `json:"data"`
	}{Data: map[string]bool{"email": true, "in_app": true}}}, nil
}

// GetTypes returns all notification types
func (h *NotificationHandler) GetTypes(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*struct {
	Body struct {
		Data []string `json:"data"`
	}
}, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid Token")
	}
	types := []string{"system", "leave", "birthday", "document", "announcement"}
	return &struct {
		Body struct {
			Data []string `json:"data"`
		}
	}{Body: struct {
		Data []string `json:"data"`
	}{Data: types}}, nil
}
