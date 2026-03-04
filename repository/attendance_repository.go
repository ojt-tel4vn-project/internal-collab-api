package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/gorm"
)

type AttendanceRepository interface {
	// Attendance records
	CreateOrUpdateAttendance(a *models.Attendance) error
	FindAttendanceByID(id uuid.UUID) (*models.Attendance, error)
	FindAttendanceByEmployeeAndMonth(employeeID uuid.UUID, month, year int) (*models.Attendance, error)
	UpdateAttendance(a *models.Attendance) error
	FindAttendances(employeeID *uuid.UUID, month, year int, status string, page, limit int) ([]models.Attendance, int64, error)
	FindAttendancesByMonthYear(month, year int) ([]models.Attendance, error)
	CountByMonthYearAndStatus(month, year int) (map[string]int, error)

	// Comments
	CreateComment(c *models.AttendanceComment) error
	FindCommentByID(id uuid.UUID) (*models.AttendanceComment, error)
	UpdateComment(c *models.AttendanceComment) error
	FindCommentsByAttendance(attendanceID uuid.UUID) ([]models.AttendanceComment, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

// --- Attendance Records ---

func (r *attendanceRepository) CreateOrUpdateAttendance(a *models.Attendance) error {
	return r.db.Save(a).Error
}

func (r *attendanceRepository) FindAttendanceByID(id uuid.UUID) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Preload("Employee").Preload("Uploader").First(&a, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepository) FindAttendanceByEmployeeAndMonth(employeeID uuid.UUID, month, year int) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Preload("Employee").
		Where("employee_id = ? AND month = ? AND year = ?", employeeID, month, year).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepository) UpdateAttendance(a *models.Attendance) error {
	return r.db.Save(a).Error
}

func (r *attendanceRepository) FindAttendances(employeeID *uuid.UUID, month, year int, status string, page, limit int) ([]models.Attendance, int64, error) {
	var records []models.Attendance
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&models.Attendance{})

	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	if month > 0 {
		query = query.Where("month = ?", month)
	}
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Preload("Employee").
		Order("year DESC, month DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&records).Error

	return records, total, err
}

func (r *attendanceRepository) FindAttendancesByMonthYear(month, year int) ([]models.Attendance, error) {
	var records []models.Attendance
	err := r.db.Where("month = ? AND year = ?", month, year).Find(&records).Error
	return records, err
}

func (r *attendanceRepository) CountByMonthYearAndStatus(month, year int) (map[string]int, error) {
	type Result struct {
		Status string
		Count  int
	}
	var results []Result
	err := r.db.Model(&models.Attendance{}).
		Select("status, COUNT(*) as count").
		Where("month = ? AND year = ?", month, year).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// --- Comments ---

func (r *attendanceRepository) CreateComment(c *models.AttendanceComment) error {
	return r.db.Create(c).Error
}

func (r *attendanceRepository) FindCommentByID(id uuid.UUID) (*models.AttendanceComment, error) {
	var c models.AttendanceComment
	err := r.db.Preload("Employee").Preload("Reviewer").First(&c, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *attendanceRepository) UpdateComment(c *models.AttendanceComment) error {
	return r.db.Save(c).Error
}

func (r *attendanceRepository) FindCommentsByAttendance(attendanceID uuid.UUID) ([]models.AttendanceComment, error) {
	var comments []models.AttendanceComment
	err := r.db.Preload("Employee").Preload("Reviewer").
		Where("attendance_id = ?", attendanceID).
		Order("created_at desc").
		Find(&comments).Error
	return comments, err
}
