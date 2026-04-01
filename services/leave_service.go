package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/leave"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
)

var (
	ErrQuotaExceeded           = errors.New("leave quota exceeded")
	ErrLeaveRequestNotFound    = errors.New("leave request not found")
	ErrLeaveTypeNotFound       = errors.New("leave type not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type LeaveService interface {
	GetLeaveTypes() ([]leave.LeaveTypeResponse, error)
	GetLeaveQuotas(employeeID uuid.UUID, year int) ([]leave.LeaveQuotaResponse, error)
	UpdateLeaveQuota(id uuid.UUID, req leave.UpdateLeaveQuotaRequest) error

	CreateLeaveRequest(employeeID uuid.UUID, req leave.CreateLeaveRequest) (*leave.LeaveRequestResponse, *leave.LeaveRequestWarning, error)
	ApproveLeaveRequest(managerID uuid.UUID, id uuid.UUID, action string, comment string) error
	CancelLeaveRequest(employeeID uuid.UUID, id uuid.UUID) error
	EmailActionLeaveRequest(token string, action string) error

	GetMyLeaveRequests(employeeID uuid.UUID, page, limit int) ([]leave.LeaveRequestResponse, int64, error)
	GetPendingLeaveRequests(managerID uuid.UUID, page, limit int) ([]leave.LeaveRequestResponse, int64, error)
	GetCompanyLeaveOverview(year, month int) (*leave.LeaveOverview, error)
}

type leaveService struct {
	repo         repository.LeaveRepository
	employeeRepo repository.EmployeeRepository
	jwtService   crypto.JWTService
}

func NewLeaveService(repo repository.LeaveRepository, employeeRepo repository.EmployeeRepository, jwtService crypto.JWTService) LeaveService {
	return &leaveService{
		repo:         repo,
		employeeRepo: employeeRepo,
		jwtService:   jwtService,
	}
}

func (s *leaveService) GetLeaveTypes() ([]leave.LeaveTypeResponse, error) {
	types, err := s.repo.FindLeaveTypes()
	if err != nil {
		return nil, err
	}

	var res []leave.LeaveTypeResponse
	for _, t := range types {
		defaultDays := 0.0
		name := strings.ToLower(t.Name)
		if strings.Contains(name, "annual leave") {
			defaultDays = 12.0
		} else if strings.Contains(name, "sick leave") {
			defaultDays = 10.0
		} else if strings.Contains(name, "compassionate") || strings.Contains(name, "bereavement") {
			defaultDays = 3.0
		} else if strings.Contains(name, "maternity") {
			defaultDays = 180.0
		} else if strings.Contains(name, "paternity") {
			defaultDays = 5.0
		} else if strings.Contains(name, "unpaid") {
			defaultDays = 30.0
		}

		res = append(res, leave.LeaveTypeResponse{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			TotalDays:   defaultDays,
		})
	}
	return res, nil
}

func (s *leaveService) GetLeaveQuotas(employeeID uuid.UUID, year int) ([]leave.LeaveQuotaResponse, error) {
	quotas, err := s.repo.FindLeaveQuotasByEmployeeAndYear(employeeID, year)
	if err != nil {
		return nil, err
	}

	var res []leave.LeaveQuotaResponse
	for _, q := range quotas {
		typeName := ""
		if q.LeaveType != nil {
			typeName = q.LeaveType.Name
		}
		
		res = append(res, leave.LeaveQuotaResponse{
			ID:            q.ID,
			EmployeeID:    q.EmployeeID,
			LeaveTypeID:   q.LeaveTypeID,
			LeaveTypeName: typeName,
			Year:          q.Year,
			TotalDays:     q.TotalDays,
			UsedDays:      q.UsedDays,
			RemainingDays: q.RemainingDays,
		})
	}
	return res, nil
}

func (s *leaveService) UpdateLeaveQuota(id uuid.UUID, req leave.UpdateLeaveQuotaRequest) error {
	if req.TotalDays < 0 {
		return errors.New("total_days must be non-negative")
	}

	quota, err := s.repo.FindLeaveQuotaByID(id)
	if err != nil {
		return err
	}
	if quota == nil {
		return errors.New("leave quota not found")
	}

	return s.repo.UpdateLeaveQuotaByID(id, req.TotalDays)
}

func (s *leaveService) CreateLeaveRequest(employeeID uuid.UUID, req leave.CreateLeaveRequest) (*leave.LeaveRequestResponse, *leave.LeaveRequestWarning, error) {
	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return nil, nil, errors.New("invalid from_date format")
	}

	todayStr := time.Now().Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)
	// Block past dates AND same-day leave (must request at least 1 day in advance)
	if !fromDate.After(today) {
		return nil, nil, errors.New("Leave start date cannot be in the past or on the same day")
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return nil, nil, errors.New("invalid to_date format")
	}

	if toDate.Before(fromDate) {
		return nil, nil, errors.New("Leave end date cannot be before start date")
	}

	// Calculate total days (simple calculation, real world would exclude weekends and holidays)
	totalDays := toDate.Sub(fromDate).Hours()/24 + 1

	leaveType, err := s.repo.FindLeaveTypeByID(req.LeaveTypeID)
	if err != nil || leaveType == nil {
		return nil, nil, ErrLeaveTypeNotFound
	}

	emp, err := s.employeeRepo.FindByID(employeeID)
	if err != nil || emp == nil {
		return nil, nil, errors.New("employee not found")
	}

	year := fromDate.Year()
	quota, err := s.repo.FindLeaveQuota(employeeID, req.LeaveTypeID, year)
	if err != nil {
		return nil, nil, err
	}

	var warning *leave.LeaveRequestWarning
	if quota != nil && quota.RemainingDays < totalDays {
		return nil, nil, errors.New("The number of leave days exceeds the limit prescribed for this leave type")
	}
	if quota == nil && totalDays > 30 {
		return nil, nil, errors.New("The number of leave days exceeds the limit prescribed for this leave type")
	}

	actionToken := uuid.New().String()

	newReq := &models.LeaveRequest{
		EmployeeID:         employeeID,
		LeaveTypeID:        req.LeaveTypeID,
		FromDate:           fromDate,
		ToDate:             toDate,
		TotalDays:          totalDays,
		Reason:             req.Reason,
		ContactDuringLeave: req.ContactDuringLeave,
		Status:             models.LeaveRequestStatusPending,
		ActionToken:        &actionToken,
	}

	err = s.repo.CreateLeaveRequest(newReq)
	if err != nil {
		return nil, nil, err
	}

	// In real world, send email to Manager here

	res := s.mapToResponse(newReq, emp, leaveType, nil)
	return res, warning, nil
}

func (s *leaveService) ApproveLeaveRequest(managerID uuid.UUID, id uuid.UUID, action string, comment string) error {
	req, err := s.repo.FindLeaveRequestByID(id)
	if err != nil || req == nil {
		return ErrLeaveRequestNotFound
	}

	if req.Status != models.LeaveRequestStatusPending {
		return ErrInvalidStatusTransition
	}

	// Verify manager
	if req.Employee == nil {
		emp, _ := s.employeeRepo.FindByID(req.EmployeeID)
		req.Employee = emp
	}

	if req.Employee.ManagerID == nil || *req.Employee.ManagerID != managerID {
		return errors.New("you are not the manager of this employee")
	}

	status := models.LeaveRequestStatusApproved
	if action == "reject" {
		status = models.LeaveRequestStatusRejected
	}

	req.Status = status
	req.ApproverID = &managerID
	req.ApproverComment = comment

	// Use transaction
	return s.repo.Transaction(func(txRepo repository.LeaveRepository) error {
		if err := txRepo.UpdateLeaveRequest(req); err != nil {
			return err
		}

		// Update quota if approved
		if status == models.LeaveRequestStatusApproved {
			year := req.FromDate.Year()
			quota, err := txRepo.FindLeaveQuota(req.EmployeeID, req.LeaveTypeID, year)
			if err != nil {
				return err
			}
			if quota != nil {
				quota.UsedDays += req.TotalDays
				if err := txRepo.UpdateLeaveQuota(quota); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *leaveService) CancelLeaveRequest(employeeID uuid.UUID, id uuid.UUID) error {
	req, err := s.repo.FindLeaveRequestByID(id)
	if err != nil || req == nil {
		return ErrLeaveRequestNotFound
	}

	if req.EmployeeID != employeeID {
		return errors.New("unauthorized")
	}

	if req.Status != models.LeaveRequestStatusPending {
		return ErrInvalidStatusTransition
	}

	todayStr := time.Now().Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)
	if req.FromDate.Before(today) {
		return errors.New("Không thể hủy đơn nghỉ phép đã quá hạn")
	}

	req.Status = models.LeaveRequestStatusCanceled
	return s.repo.UpdateLeaveRequest(req)
}

func (s *leaveService) EmailActionLeaveRequest(token string, action string) error {
	req, err := s.repo.FindLeaveRequestByActionToken(token)
	if err != nil || req == nil {
		return errors.New("invalid or expired token")
	}

	if req.Status != models.LeaveRequestStatusPending {
		return ErrInvalidStatusTransition
	}

	status := models.LeaveRequestStatusApproved
	if action == "reject" {
		status = models.LeaveRequestStatusRejected
	}

	req.Status = status

	if req.Employee != nil && req.Employee.ManagerID != nil {
		req.ApproverID = req.Employee.ManagerID
	}

	// Invalidate token
	req.ActionToken = nil

	return s.repo.Transaction(func(txRepo repository.LeaveRepository) error {
		if err := txRepo.UpdateLeaveRequest(req); err != nil {
			return err
		}

		if status == models.LeaveRequestStatusApproved {
			year := req.FromDate.Year()
			quota, err := txRepo.FindLeaveQuota(req.EmployeeID, req.LeaveTypeID, year)
			if err != nil {
				return err
			}
			if quota != nil {
				quota.UsedDays += req.TotalDays
				if err := txRepo.UpdateLeaveQuota(quota); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *leaveService) GetMyLeaveRequests(employeeID uuid.UUID, page, limit int) ([]leave.LeaveRequestResponse, int64, error) {
	emp, err := s.employeeRepo.FindByID(employeeID)
	if err != nil || emp == nil {
		return nil, 0, errors.New("employee not found")
	}

	reqs, total, err := s.repo.FindLeaveRequestsByEmployee(employeeID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	res := make([]leave.LeaveRequestResponse, 0, len(reqs))
	for _, req := range reqs {
		res = append(res, *s.mapToResponse(&req, emp, req.LeaveType, req.Approver))
	}
	return res, total, nil
}

func (s *leaveService) GetPendingLeaveRequests(managerID uuid.UUID, page, limit int) ([]leave.LeaveRequestResponse, int64, error) {
	reqs, total, err := s.repo.FindPendingLeaveRequestsByManager(managerID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	res := make([]leave.LeaveRequestResponse, 0)
	for _, req := range reqs {
		res = append(res, *s.mapToResponse(&req, req.Employee, req.LeaveType, req.Approver))
	}
	return res, total, nil
}

func (s *leaveService) GetCompanyLeaveOverview(year, month int) (*leave.LeaveOverview, error) {
	reqs, err := s.repo.FindAllLeaveRequestsOverview(year, month)
	if err != nil {
		return nil, err
	}

	overview := &leave.LeaveOverview{}
	overview.TotalRequests = len(reqs)

	today := time.Now().Truncate(24 * time.Hour)

	for _, req := range reqs {
		switch req.Status {
		case models.LeaveRequestStatusPending:
			overview.Pending++
		case models.LeaveRequestStatusApproved:
			overview.Approved++
			// Check if on leave today
			if (req.FromDate.Equal(today) || req.FromDate.Before(today)) &&
				(req.ToDate.Equal(today) || req.ToDate.After(today)) {
				overview.EmployeesOnLeaveToday++
			}
			// Check if upcoming
			if req.FromDate.After(today) {
				overview.UpcomingLeaves = append(overview.UpcomingLeaves, struct {
					Employee string `json:"employee"`
					FromDate string `json:"from_date"`
					ToDate   string `json:"to_date"`
				}{
					Employee: req.Employee.FullName,
					FromDate: req.FromDate.Format("2006-01-02"),
					ToDate:   req.ToDate.Format("2006-01-02"),
				})
			}
		case models.LeaveRequestStatusRejected:
			overview.Rejected++
		}
	}

	if overview.UpcomingLeaves == nil {
		overview.UpcomingLeaves = make([]struct {
			Employee string "json:\"employee\""
			FromDate string "json:\"from_date\""
			ToDate   string "json:\"to_date\""
		}, 0)
	}

	return overview, nil
}

func (s *leaveService) mapToResponse(req *models.LeaveRequest, emp *models.Employee, t *models.LeaveType, approver *models.Employee) *leave.LeaveRequestResponse {
	todayStr := time.Now().Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)
	isOverdue := false
	if req.Status == models.LeaveRequestStatusPending && req.FromDate.Before(today) {
		isOverdue = true
	}

	res := &leave.LeaveRequestResponse{
		ID:                 req.ID,
		FromDate:           req.FromDate.Format("2006-01-02"),
		ToDate:             req.ToDate.Format("2006-01-02"),
		TotalDays:          req.TotalDays,
		Reason:             req.Reason,
		ContactDuringLeave: req.ContactDuringLeave,
		Status:             req.Status,
		IsOverdue:          isOverdue,
		ApproverComment:    req.ApproverComment,
		SubmittedAt:        req.SubmittedAt,
	}

	if emp != nil {
		res.Employee = leave.EmployeeResponse{
			ID:        emp.ID,
			FullName:  emp.FullName,
			Email:     emp.Email,
			Position:  emp.Position,
			AvatarUrl: emp.AvatarUrl,
		}
	}

	if t != nil {
		res.LeaveType = leave.LeaveTypeResponse{
			ID:   t.ID,
			Name: t.Name,
		}
	}

	if approver != nil {
		res.Approver = &leave.EmployeeResponse{
			ID:        approver.ID,
			FullName:  approver.FullName,
			Email:     approver.Email,
			Position:  approver.Position,
			AvatarUrl: approver.AvatarUrl,
		}
	}

	return res
}
