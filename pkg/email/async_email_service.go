package email

import (
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"go.uber.org/zap"
)

type asyncEmailService struct {
	baseService EmailService
	jobQueue    chan func() error
}

// NewAsyncEmailService wraps an existing EmailService to execute its methods asynchronously
// It creates a buffered queue and launches background workers
func NewAsyncEmailService(baseService EmailService, workerCount int) EmailService {
	s := &asyncEmailService{
		baseService: baseService,
		jobQueue:    make(chan func() error, 1000), // Queue buffer up to 1000 emails
	}

	// Start background workers
	for i := 0; i < workerCount; i++ {
		go s.worker()
	}

	return s
}

func (s *asyncEmailService) worker() {
	// Constantly read from the queue and send emails in the background
	for job := range s.jobQueue {
		if err := job(); err != nil {
			logger.Error("Async email job failed", zap.Error(err))
		}
	}
}

func (s *asyncEmailService) enqueue(job func() error) error {
	select {
	case s.jobQueue <- job:
		return nil // Successfully pushed to the queue
	default:
		// Queue is full, executing synchronously as fallback
		logger.Warn("Email queue is full, executing synchronously as fallback")
		return job()
	}
}

// Implement EmailService interface exactly matching the original methods

func (s *asyncEmailService) SendWelcomeEmail(to, name, tempPassword string) error {
	job := func() error {
		return s.baseService.SendWelcomeEmail(to, name, tempPassword)
	}
	_ = s.enqueue(job)
	return nil
}

func (s *asyncEmailService) SendPasswordResetEmail(to, name, resetLink string) error {
	job := func() error {
		return s.baseService.SendPasswordResetEmail(to, name, resetLink)
	}
	_ = s.enqueue(job)
	return nil
}

func (s *asyncEmailService) SendPasswordChangedEmail(to, name string) error {
	job := func() error {
		return s.baseService.SendPasswordChangedEmail(to, name)
	}
	_ = s.enqueue(job)
	return nil
}

func (s *asyncEmailService) SendBirthdayWish(to, name string) error {
	job := func() error {
		return s.baseService.SendBirthdayWish(to, name)
	}
	_ = s.enqueue(job)
	return nil
}
