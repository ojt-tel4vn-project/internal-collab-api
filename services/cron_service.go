package services

import (
	"time"

	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/email"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CronService interface {
	Start()
	Stop()
}

type cronService struct {
	cron         *cron.Cron
	employeeRepo repository.EmployeeRepository
	emailService email.EmailService
}

func NewCronService(employeeRepo repository.EmployeeRepository, emailService email.EmailService) CronService {
	c := cron.New()
	s := &cronService{
		cron:         c,
		employeeRepo: employeeRepo,
		emailService: emailService,
	}

	// Add jobs
	// Run every day at 9:00 AM
	// "0 9 * * *" is standard cron syntax (Minute Hour DayOfMonth Month DayOfWeek)
	_, err := c.AddFunc("0 9 * * *", s.checkBirthdays)
	if err != nil {
		logger.Error("Failed to add birthday check job", zap.Error(err))
	}

	return s
}

func (s *cronService) Start() {
	s.cron.Start()
	logger.Info("Cron service started")
}

func (s *cronService) Stop() {
	s.cron.Stop()
	logger.Info("Cron service stopped")
}

func (s *cronService) checkBirthdays() {
	today := time.Now()
	month := int(today.Month())
	day := today.Day()

	logger.Info("Checking birthdays for", zap.Int("month", month), zap.Int("day", day))

	employees, err := s.employeeRepo.FindEmployeesByBirthday(month, day)
	if err != nil {
		logger.Error("Failed to fetch employees with birthdays", zap.Error(err))
		return
	}

	if len(employees) == 0 {
		logger.Info("No birthdays today")
		return
	}

	for _, emp := range employees {
		if s.emailService != nil {
			err := s.emailService.SendBirthdayWish(emp.Email, emp.FullName)
			if err != nil {
				logger.Error("Failed to send birthday email", zap.String("email", emp.Email), zap.Error(err))
			} else {
				logger.Info("Sent birthday wish", zap.String("email", emp.Email))
			}
		}
	}
}
