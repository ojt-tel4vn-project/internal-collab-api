package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	empdto "github.com/ojt-tel4vn-project/internal-collab-api/dtos/employee"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
	"github.com/ojt-tel4vn-project/internal-collab-api/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── helpers ──────────────────────────────────────────────────

func newEmployeeService(empRepo *mocks.MockEmployeeRepository, pw *mocks.MockPasswordService, email *mocks.MockEmailService) services.EmployeeService {
	appCfg := &mocks.MockAppConfigRepository{}
	return services.NewEmployeeService(empRepo, pw, email, appCfg)
}

func sampleCreateRequest() *empdto.CreateEmployeeRequest {
	return &empdto.CreateEmployeeRequest{
		Email:       "new.hire@company.com",
		FirstName:   "New",
		LastName:    "Hire",
		DateOfBirth: "1995-06-15",
		Position:    "Backend Developer",
		JoinDate:    "2026-03-01",
	}
}

// ═══════════════════════════════════════════════════════════════
// CreateEmployee Tests
// ═══════════════════════════════════════════════════════════════

func TestCreateEmployee_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	req := sampleCreateRequest()

	empRepo.On("FindByEmail", req.Email).Return(nil, errors.New("not found")) // email not taken
	pw.On("HashPassword", mock.AnythingOfType("string")).Return("$2a$10$hashedtemp", nil)
	empRepo.On("Create", mock.AnythingOfType("*models.Employee")).Return(nil)
	email.On("SendWelcomeEmail", req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.CreateEmployee(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "new.hire@company.com", resp.Employee.Email)
	assert.Equal(t, "pending", resp.Employee.Status)
	assert.NotEmpty(t, resp.TemporaryPassword)
	email.AssertCalled(t, "SendWelcomeEmail", req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string"))
}

func TestCreateEmployee_DuplicateEmail(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	req := sampleCreateRequest()
	existing := &models.Employee{ID: uuid.New(), Email: req.Email}

	empRepo.On("FindByEmail", req.Email).Return(existing, nil) // email already taken

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.CreateEmployee(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	empRepo.AssertNotCalled(t, "Create")
	email.AssertNotCalled(t, "SendWelcomeEmail")
}

func TestCreateEmployee_InvalidDateOfBirth(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	req := sampleCreateRequest()
	req.DateOfBirth = "not-a-date"

	empRepo.On("FindByEmail", req.Email).Return(nil, errors.New("not found"))
	pw.On("HashPassword", mock.AnythingOfType("string")).Return("$2a$10$hashedtemp", nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.CreateEmployee(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "date of birth")
}

func TestCreateEmployee_EmailFailShouldNotFail(t *testing.T) {
	// Email failure should NOT fail the whole operation — employee still created
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	emailSvc := &mocks.MockEmailService{}

	req := sampleCreateRequest()

	empRepo.On("FindByEmail", req.Email).Return(nil, errors.New("not found"))
	pw.On("HashPassword", mock.AnythingOfType("string")).Return("$2a$10$hashedtemp", nil)
	empRepo.On("Create", mock.AnythingOfType("*models.Employee")).Return(nil)
	emailSvc.On("SendWelcomeEmail", req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(errors.New("smtp timeout")) // email fails

	svc := newEmployeeService(empRepo, pw, emailSvc)
	resp, err := svc.CreateEmployee(req)

	assert.NoError(t, err) // must succeed even if email fails
	assert.NotNil(t, resp) // employee created
}

// ═══════════════════════════════════════════════════════════════
// GetAllEmployees Tests
// ═══════════════════════════════════════════════════════════════

func TestGetAllEmployees_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	employees := []models.Employee{
		{ID: uuid.New(), Email: "alice@co.com", FullName: "Alice", Position: "PM", Status: models.StatusActive},
		{ID: uuid.New(), Email: "bob@co.com", FullName: "Bob", Position: "Dev", Status: models.StatusActive},
	}
	empRepo.On("FindAll").Return(employees, nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.GetAllEmployees()

	assert.NoError(t, err)
	assert.Len(t, resp.Employees, 2)
	assert.Equal(t, 2, resp.Total)
}

func TestGetAllEmployees_Empty(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	empRepo.On("FindAll").Return([]models.Employee{}, nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.GetAllEmployees()

	assert.NoError(t, err)
	assert.Len(t, resp.Employees, 0)
	assert.Equal(t, 0, resp.Total)
}

// ═══════════════════════════════════════════════════════════════
// UpdateEmployee Tests
// ═══════════════════════════════════════════════════════════════

func TestUpdateEmployee_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	emp := &models.Employee{
		ID:       uuid.New(),
		FullName: "Old Name",
		Position: "Old Position",
		Status:   models.StatusActive,
	}
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	empRepo.On("Update", emp).Return(nil)

	newPosition := "Senior Developer"
	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.UpdateEmployee(emp.ID, &empdto.UpdateEmployeeRequest{
		Position: &newPosition,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Senior Developer", emp.Position)
}

func TestUpdateEmployee_NotFound(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	id := uuid.New()
	empRepo.On("FindByID", id).Return(nil, errors.New("not found"))

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.UpdateEmployee(id, &empdto.UpdateEmployeeRequest{})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ═══════════════════════════════════════════════════════════════
// DeleteEmployee Tests
// ═══════════════════════════════════════════════════════════════

func TestDeleteEmployee_Success(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	emp := &models.Employee{ID: uuid.New(), Status: models.StatusActive}
	empRepo.On("FindByID", emp.ID).Return(emp, nil)
	empRepo.On("Delete", emp.ID).Return(nil)

	svc := newEmployeeService(empRepo, pw, email)
	err := svc.DeleteEmployee(emp.ID)

	assert.NoError(t, err)
}

func TestDeleteEmployee_NotFound(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	id := uuid.New()
	empRepo.On("FindByID", id).Return(nil, errors.New("not found"))

	svc := newEmployeeService(empRepo, pw, email)
	err := svc.DeleteEmployee(id)

	assert.Error(t, err)
}

// ═══════════════════════════════════════════════════════════════
// GetTodayBirthdays Tests
// ═══════════════════════════════════════════════════════════════

func TestGetTodayBirthdays_HasBirthdays(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	now := time.Now()
	employees := []models.Employee{
		{
			ID:          uuid.New(),
			FullName:    "Birthday Person",
			Email:       "birthday@co.com",
			Position:    "Dev",
			DateOfBirth: time.Date(1990, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		},
	}
	empRepo.On("FindEmployeesByBirthday", int(now.Month()), now.Day()).Return(employees, nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.GetTodayBirthdays()

	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "Found birthdays today", resp.Message)
}

func TestGetTodayBirthdays_NoBirthdays(t *testing.T) {
	empRepo := &mocks.MockEmployeeRepository{}
	pw := &mocks.MockPasswordService{}
	email := &mocks.MockEmailService{}

	now := time.Now()
	empRepo.On("FindEmployeesByBirthday", int(now.Month()), now.Day()).Return([]models.Employee{}, nil)

	svc := newEmployeeService(empRepo, pw, email)
	resp, err := svc.GetTodayBirthdays()

	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, "No birthdays today", resp.Message)
}
