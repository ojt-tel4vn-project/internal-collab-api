package comment

import (
	"time"

	"github.com/google/uuid"
)

// CreateCommentRequest payload for creating a comment
type CreateCommentRequest struct {
	Content  string     `json:"content" required:"true" doc:"Comment text"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" doc:"Parent comment ID for replies"`
}

// CommentAuthor slim author info embedded in responses
type CommentAuthor struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarUrl string    `json:"avatar_url"`
}

// CommentItem is a single comment in a list
type CommentItem struct {
	ID         uuid.UUID     `json:"id"`
	DocumentID uuid.UUID     `json:"document_id"`
	Author     CommentAuthor `json:"author"`
	Content    string        `json:"content"`
	ParentID   *uuid.UUID    `json:"parent_id,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// ListCommentsResponse response for listing comments of a document
type ListCommentsResponse struct {
	Comments []CommentItem `json:"comments"`
	Total    int           `json:"total"`
}

// CreateCommentResponse response after creating a comment
type CreateCommentResponse struct {
	Message string      `json:"message"`
	Comment CommentItem `json:"comment"`
}

// DeleteCommentResponse response after deleting a comment
type DeleteCommentResponse struct {
	Message string `json:"message"`
}
