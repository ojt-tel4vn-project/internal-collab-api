package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	attendancedto "github.com/ojt-tel4vn-project/internal-collab-api/dtos/attendance"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

var (
	ErrAttendanceNotFound = errors.New("attendance record not found")
	ErrCommentNotFound    = errors.New("attendance comment not found")
	ErrAlreadyConfirmed   = errors.New("attendance already confirmed or auto-confirmed")
	ErrUnauthorizedAccess = errors.New("unauthorized access to attendance record")
)

// AttendanceService defines the business logic for attendance management
type AttendanceService interface {
	// HR operations
	UploadAttendance(uploaderID uuid.UUID, month, year int, csvContent string) ([]attendancedto.AttendanceResponse, error)
	GetAttendanceSummary(month, year int) (*attendancedto.AttendanceSummary, error)
	GetAttendanceConfig() (*attendancedto.AttendanceConfig, error)
	UpdateAttendanceConfig(req attendancedto.UpdateAttendanceConfigRequest) error
	ReviewComment(hrID uuid.UUID, commentID uuid.UUID, req attendancedto.ReviewCommentRequest) error

	// Employee & HR operations
	ListAttendances(employeeID *uuid.UUID, month, year int, status string, page, limit int) ([]attendancedto.AttendanceResponse, int64, error)
	GetAttendanceByID(id uuid.UUID) (*attendancedto.AttendanceResponse, error)
	ConfirmAttendance(employeeID uuid.UUID, attendanceID uuid.UUID, req attendancedto.ConfirmAttendanceRequest) error
	AddComment(employeeID uuid.UUID, attendanceID uuid.UUID, req attendancedto.AddCommentRequest) (*attendancedto.AttendanceCommentResponse, error)

	// Cron
	AutoConfirmOverdueAttendances() error
}

type attendanceService struct {
	repo          repository.AttendanceRepository
	employeeRepo  repository.EmployeeRepository
	appConfigRepo repository.AppConfigRepository
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	employeeRepo repository.EmployeeRepository,
	appConfigRepo repository.AppConfigRepository,
) AttendanceService {
	return &attendanceService{
		repo:          repo,
		employeeRepo:  employeeRepo,
		appConfigRepo: appConfigRepo,
	}
}

// UploadAttendance parses a CSV file and upserts attendance records for the given month/year.
// CSV format: employee_code, day1, day2, ..., day31
// Values: present | absent | late | leave (empty = not counted)
func (s *attendanceService) UploadAttendance(uploaderID uuid.UUID, month, year int, csvContent string) ([]attendancedto.AttendanceResponse, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV format: %w", err)
	}

	if len(records) < 2 {
		return nil, errors.New("CSV must contain a header row and at least one data row")
	}

	// Map to store attendance data by employee ID
	attendanceMap := make(map[string]models.AttendanceData)
	employeeCache := make(map[string]*models.Employee)

	// Parse CSV records
	// Expected format: ID Người, Tên, Bộ phận, Ngày, Thời gian biểu, Tình trạng chuyên cần, Vào, Ra
	for i, row := range records {
		// Skip header row and detail rows (rows starting with "Thời gian vào:")
		if i == 0 || len(row) < 6 {
			continue
		}

		// Skip detail/summary rows
		firstCol := strings.TrimSpace(row[0])
		if firstCol == "" || strings.HasPrefix(firstCol, "Thời gian") || firstCol == "Chi Tiết Chấm Công" {
			continue
		}

		employeeCode := firstCol
		dateStr := strings.TrimSpace(row[3])     // Ngày: 2025-07-28
		statusStr := strings.TrimSpace(row[5])   // Tình trạng: bình thường, Muộn, Về sớm, Vắng mặt
		checkInStr := strings.TrimSpace(row[6])  // Vào: 08:19:35 or -
		checkOutStr := strings.TrimSpace(row[7]) // Ra: 19:26:14 or -

		// Parse date to get day number
		dateParts := strings.Split(dateStr, "-")
		if len(dateParts) != 3 {
			continue
		}
		day, err := strconv.Atoi(dateParts[2])
		if err != nil || day < 1 || day > 31 {
			continue
		}

		// Find or cache employee
		emp, exists := employeeCache[employeeCode]
		if !exists {
			emp, err = s.employeeRepo.FindByEmployeeCode(employeeCode)
			if err != nil || emp == nil {
				logger.Warn("Employee not found during attendance upload", zap.String("code", employeeCode))
				continue
			}
			employeeCache[employeeCode] = emp
		}

		// Initialize attendance data for this employee if not exists
		if attendanceMap[employeeCode] == nil {
			attendanceMap[employeeCode] = make(models.AttendanceData)
		}

		// Map status from Vietnamese to system status
		var status models.DayStatus
		switch strings.ToLower(statusStr) {
		case "bình thường":
			status = models.DayStatusPresent
		case "muộn":
			status = models.DayStatusLate
		case "vắng mặt":
			status = models.DayStatusAbsent
		case "về sớm":
			status = models.DayStatusPresent // Consider early leave as present but can be tracked
		default:
			status = models.DayStatusPresent
		}

		// Calculate work hours if check-in and check-out are available
		var workHours float64
		if checkInStr != "-" && checkOutStr != "-" && checkInStr != "" && checkOutStr != "" {
			workHours = calculateWorkHours(checkInStr, checkOutStr)
		}

		// Store day attendance detail
		dayKey := strconv.Itoa(day)
		attendanceMap[employeeCode][dayKey] = models.DayAttendanceDetail{
			Status:       status,
			CheckInTime:  checkInStr,
			CheckOutTime: checkOutStr,
			WorkHours:    workHours,
		}
	}

	// Create or update attendance records
	var results []attendancedto.AttendanceResponse
	for employeeCode, data := range attendanceMap {
		emp := employeeCache[employeeCode]

		// Calculate totals
		var present, absent, late int32
		for _, detail := range data {
			switch detail.Status {
			case models.DayStatusPresent:
				present++
			case models.DayStatusAbsent:
				absent++
			case models.DayStatusLate:
				late++
			}
		}

		// Upsert
		existing, _ := s.repo.FindAttendanceByEmployeeAndMonth(emp.ID, month, year)
		if existing != nil {
			existing.AttendanceData = data
			existing.TotalDaysPresent = present
			existing.TotalDaysAbsent = absent
			existing.TotalDaysLate = late
			existing.UploadedBy = &uploaderID
			if err := s.repo.UpdateAttendance(existing); err != nil {
				return nil, fmt.Errorf("failed to update attendance for %s: %w", employeeCode, err)
			}
			results = append(results, s.mapToResponse(existing))
		} else {
			newRecord := &models.Attendance{
				EmployeeID:       emp.ID,
				Month:            int32(month),
				Year:             int32(year),
				AttendanceData:   data,
				TotalDaysPresent: present,
				TotalDaysAbsent:  absent,
				TotalDaysLate:    late,
				Status:           models.AttendanceStatusPending,
				UploadedBy:       &uploaderID,
			}
			if err := s.repo.CreateOrUpdateAttendance(newRecord); err != nil {
				return nil, fmt.Errorf("failed to create attendance for %s: %w", employeeCode, err)
			}
			newRecord.Employee = emp
			results = append(results, s.mapToResponse(newRecord))
		}
	}

	return results, nil
}

// calculateWorkHours calculates work hours from check-in and check-out time strings (HH:MM:SS)
func calculateWorkHours(checkIn, checkOut string) float64 {
	parseTime := func(timeStr string) (int, int, int) {
		parts := strings.Split(timeStr, ":")
		if len(parts) != 3 {
			return 0, 0, 0
		}
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		s, _ := strconv.Atoi(parts[2])
		return h, m, s
	}

	inH, inM, inS := parseTime(checkIn)
	outH, outM, outS := parseTime(checkOut)

	inSeconds := inH*3600 + inM*60 + inS
	outSeconds := outH*3600 + outM*60 + outS

	diffSeconds := outSeconds - inSeconds
	if diffSeconds < 0 {
		diffSeconds += 24 * 3600 // Handle overnight shift
	}

	return float64(diffSeconds) / 3600.0
}

func (s *attendanceService) ListAttendances(employeeID *uuid.UUID, month, year int, status string, page, limit int) ([]attendancedto.AttendanceResponse, int64, error) {
	records, total, err := s.repo.FindAttendances(employeeID, month, year, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var res []attendancedto.AttendanceResponse
	for i := range records {
		res = append(res, s.mapToResponse(&records[i]))
	}
	return res, total, nil
}

func (s *attendanceService) GetAttendanceByID(id uuid.UUID) (*attendancedto.AttendanceResponse, error) {
	a, err := s.repo.FindAttendanceByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrAttendanceNotFound
	}
	r := s.mapToResponse(a)
	return &r, nil
}

func (s *attendanceService) ConfirmAttendance(employeeID uuid.UUID, attendanceID uuid.UUID, req attendancedto.ConfirmAttendanceRequest) error {
	a, err := s.repo.FindAttendanceByID(attendanceID)
	if err != nil || a == nil {
		return ErrAttendanceNotFound
	}

	// Only the record owner can confirm
	if a.EmployeeID != employeeID {
		return ErrUnauthorizedAccess
	}

	if a.Status != models.AttendanceStatusPending {
		return ErrAlreadyConfirmed
	}

	now := time.Now()
	a.Status = models.AttendanceStatus(req.Status)
	a.ConfirmedAt = &now

	return s.repo.UpdateAttendance(a)
}

func (s *attendanceService) AddComment(employeeID uuid.UUID, attendanceID uuid.UUID, req attendancedto.AddCommentRequest) (*attendancedto.AttendanceCommentResponse, error) {
	a, err := s.repo.FindAttendanceByID(attendanceID)
	if err != nil || a == nil {
		return nil, ErrAttendanceNotFound
	}

	if a.EmployeeID != employeeID {
		return nil, ErrUnauthorizedAccess
	}

	comment := &models.AttendanceComment{
		AttendanceID: attendanceID,
		EmployeeID:   employeeID,
		Comment:      req.Comment,
		DayNumber:    int32(req.DayNumber),
		Status:       models.CommentStatusPending,
	}

	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}

	emp, _ := s.employeeRepo.FindByID(employeeID)
	comment.Employee = emp

	resp := s.mapCommentToResponse(comment)
	return &resp, nil
}

func (s *attendanceService) ReviewComment(hrID uuid.UUID, commentID uuid.UUID, req attendancedto.ReviewCommentRequest) error {
	c, err := s.repo.FindCommentByID(commentID)
	if err != nil || c == nil {
		return ErrCommentNotFound
	}

	now := time.Now()
	c.HRResponse = req.HRResponse
	c.Status = models.AttendanceCommentStatus(req.Status)
	c.ReviewedBy = &hrID
	c.ReviewedAt = &now

	return s.repo.UpdateComment(c)
}

func (s *attendanceService) GetAttendanceSummary(month, year int) (*attendancedto.AttendanceSummary, error) {
	counts, err := s.repo.CountByMonthYearAndStatus(month, year)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, v := range counts {
		total += v
	}

	return &attendancedto.AttendanceSummary{
		TotalEmployees: total,
		Confirmed:      counts[string(models.AttendanceStatusConfirmed)],
		Pending:        counts[string(models.AttendanceStatusPending)],
		AutoConfirmed:  counts[string(models.AttendanceStatusAutoConfirmed)],
	}, nil
}

// GetAttendanceConfig reads attendance config from AppConfig (key-value store)
func (s *attendanceService) GetAttendanceConfig() (*attendancedto.AttendanceConfig, error) {
	deadlineDays := 7
	autoConfirm := true
	reminderDays := 2

	if v, err := s.appConfigRepo.Get("attendance_confirm_deadline_days"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			deadlineDays = n
		}
	}
	if v, err := s.appConfigRepo.Get("attendance_auto_confirm_enabled"); err == nil {
		autoConfirm = v == "true"
	}
	if v, err := s.appConfigRepo.Get("attendance_reminder_before_deadline_days"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			reminderDays = n
		}
	}

	return &attendancedto.AttendanceConfig{
		ConfirmationDeadlineDays:   deadlineDays,
		AutoConfirmEnabled:         autoConfirm,
		ReminderBeforeDeadlineDays: reminderDays,
	}, nil
}

// UpdateAttendanceConfig writes attendance config values to AppConfig
func (s *attendanceService) UpdateAttendanceConfig(req attendancedto.UpdateAttendanceConfigRequest) error {
	if err := s.appConfigRepo.Set("attendance_confirm_deadline_days", strconv.Itoa(req.ConfirmationDeadlineDays)); err != nil {
		return err
	}
	autoConfirmStr := "false"
	if req.AutoConfirmEnabled {
		autoConfirmStr = "true"
	}
	if err := s.appConfigRepo.Set("attendance_auto_confirm_enabled", autoConfirmStr); err != nil {
		return err
	}
	return s.appConfigRepo.Set("attendance_reminder_before_deadline_days", strconv.Itoa(req.ReminderBeforeDeadlineDays))
}

// AutoConfirmOverdueAttendances is called by cron to auto-confirm pending records past deadline
func (s *attendanceService) AutoConfirmOverdueAttendances() error {
	cfg, err := s.GetAttendanceConfig()
	if err != nil {
		return err
	}

	if !cfg.AutoConfirmEnabled {
		logger.Info("Auto-confirm disabled, skipping")
		return nil
	}

	now := time.Now()
	// Check records from the previous month
	targetMonth := int(now.Month()) - 1
	targetYear := now.Year()
	if targetMonth == 0 {
		targetMonth = 12
		targetYear--
	}

	records, _, err := s.repo.FindAttendances(nil, targetMonth, targetYear, string(models.AttendanceStatusPending), 1, 1000)
	if err != nil {
		return err
	}

	deadline := time.Date(now.Year(), now.Month(), cfg.ConfirmationDeadlineDays, 23, 59, 59, 0, now.Location())
	if now.After(deadline) {
		for i := range records {
			confirmed := time.Now()
			records[i].Status = models.AttendanceStatusAutoConfirmed
			records[i].ConfirmedAt = &confirmed
			if err := s.repo.UpdateAttendance(&records[i]); err != nil {
				logger.Error("Failed to auto-confirm attendance", zap.String("id", records[i].ID.String()), zap.Error(err))
			}
		}
		logger.Info("Auto-confirmed overdue attendances", zap.Int("count", len(records)))
	}

	return nil
}

// ---- Mappers ----

func (s *attendanceService) mapToResponse(a *models.Attendance) attendancedto.AttendanceResponse {
	resp := attendancedto.AttendanceResponse{
		ID:               a.ID,
		Month:            a.Month,
		Year:             a.Year,
		AttendanceData:   a.AttendanceData,
		TotalDaysPresent: a.TotalDaysPresent,
		TotalDaysAbsent:  a.TotalDaysAbsent,
		TotalDaysLate:    a.TotalDaysLate,
		Status:           a.Status,
		ConfirmedAt:      a.ConfirmedAt,
		UploadedAt:       a.UploadedAt,
	}
	if a.Employee != nil {
		resp.Employee = attendancedto.EmployeeRef{
			ID:       a.Employee.ID,
			FullName: a.Employee.FullName,
			Email:    a.Employee.Email,
			Position: a.Employee.Position,
		}
	}
	return resp
}

func (s *attendanceService) mapCommentToResponse(c *models.AttendanceComment) attendancedto.AttendanceCommentResponse {
	resp := attendancedto.AttendanceCommentResponse{
		ID:           c.ID,
		AttendanceID: c.AttendanceID,
		Comment:      c.Comment,
		DayNumber:    c.DayNumber,
		Status:       c.Status,
		HRResponse:   c.HRResponse,
		ReviewedAt:   c.ReviewedAt,
		CreatedAt:    c.CreatedAt,
	}
	if c.Employee != nil {
		resp.Employee = attendancedto.EmployeeRef{
			ID:       c.Employee.ID,
			FullName: c.Employee.FullName,
			Email:    c.Employee.Email,
		}
	}
	return resp
}
