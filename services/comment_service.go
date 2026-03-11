package services

import (
	commentDTO "github.com/ojt-tel4vn-project/internal-collab-api/dtos/comment"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CommentService interface {
	CreateComment(attendanceID, authorID uuid.UUID, req *commentDTO.CreateCommentRequest) (*commentDTO.CreateCommentResponse, error)
	GetCommentsByAttendance(attendanceID uuid.UUID) (*commentDTO.ListCommentsResponse, error)
	MarkRead(commentID, requestorID uuid.UUID) (*commentDTO.MarkReadResponse, error)
	DeleteComment(commentID, requestorID uuid.UUID, isHR bool) error
}

type commentServiceImpl struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentServiceImpl{repo: repo}
}

func toCommentItem(c models.Comment) commentDTO.CommentItem {
	author := commentDTO.CommentAuthor{}
	if c.Author != nil {
		author.ID = c.Author.ID
		author.FullName = c.Author.FullName
		author.AvatarUrl = c.Author.AvatarUrl
	}
	return commentDTO.CommentItem{
		ID:           c.ID,
		AttendanceID: c.AttendanceID,
		Author:       author,
		Content:      c.Content,
		IsRead:       c.IsRead,
		ParentID:     c.ParentID,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func (s *commentServiceImpl) CreateComment(attendanceID, authorID uuid.UUID, req *commentDTO.CreateCommentRequest) (*commentDTO.CreateCommentResponse, error) {
	if req.Content == "" {
		return nil, response.BadRequest("Comment content cannot be empty")
	}

	comment := &models.Comment{
		AttendanceID: attendanceID,
		AuthorID:     authorID,
		Content:      req.Content,
		ParentID:     req.ParentID,
		IsRead:       false,
	}

	if err := s.repo.Create(comment); err != nil {
		logger.Error("CreateComment: DB error", zap.Error(err))
		return nil, response.InternalServerError("Failed to create comment")
	}

	// Reload with author info
	created, err := s.repo.FindByID(comment.ID)
	if err != nil {
		logger.Warn("CreateComment: reload failed", zap.Error(err))
		created = comment
	}

	return &commentDTO.CreateCommentResponse{
		Message: "Comment created successfully",
		Comment: toCommentItem(*created),
	}, nil
}

func (s *commentServiceImpl) GetCommentsByAttendance(attendanceID uuid.UUID) (*commentDTO.ListCommentsResponse, error) {
	comments, err := s.repo.FindByAttendanceID(attendanceID)
	if err != nil {
		logger.Error("GetCommentsByAttendance: DB error", zap.Error(err))
		return nil, response.InternalServerError("Failed to fetch comments")
	}

	items := make([]commentDTO.CommentItem, len(comments))
	for i, c := range comments {
		items[i] = toCommentItem(c)
	}

	return &commentDTO.ListCommentsResponse{
		Comments: items,
		Total:    len(items),
	}, nil
}

// MarkRead marks a comment as read. Only HR/manager or the attendance owner's manager can do this.
func (s *commentServiceImpl) MarkRead(commentID, requestorID uuid.UUID) (*commentDTO.MarkReadResponse, error) {
	_, err := s.repo.FindByID(commentID)
	if err != nil {
		return nil, response.NotFound("Comment not found")
	}

	if err := s.repo.MarkRead(commentID); err != nil {
		logger.Error("MarkRead: DB error", zap.Error(err))
		return nil, response.InternalServerError("Failed to mark comment as read")
	}

	logger.Info("Comment marked as read", zap.String("comment_id", commentID.String()))
	return &commentDTO.MarkReadResponse{Message: "Comment marked as read"}, nil
}

func (s *commentServiceImpl) DeleteComment(commentID, requestorID uuid.UUID, isHR bool) error {
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return response.NotFound("Comment not found")
	}

	// Only the author or HR/admin can delete
	if !isHR && comment.AuthorID != requestorID {
		return response.Forbidden("You can only delete your own comments")
	}

	if err := s.repo.Delete(commentID); err != nil {
		logger.Error("DeleteComment: DB error", zap.Error(err))
		return response.InternalServerError("Failed to delete comment")
	}

	logger.Info("Comment deleted", zap.String("comment_id", commentID.String()))
	return nil
}
