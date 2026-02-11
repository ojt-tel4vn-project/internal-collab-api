package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	auditlog "github.com/ojt-tel4vn-project/internal-collab-api/dtos/audit_log"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
	"gorm.io/datatypes"

	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
)

type AuditLogService interface {
	Log(actorID *uuid.UUID, action, entityType string, entityID uuid.UUID, oldPayload, newPayload interface{}, description string)
	GetLogs(filter *auditlog.AuditLogFilter) (*auditlog.ListAuditLogResponse, error)
	ExportLogs(filter *auditlog.AuditLogFilter) ([]byte, string, error)
}

type auditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

func (s *auditLogService) Log(actorID *uuid.UUID, action, entityType string, entityID uuid.UUID, oldPayload, newPayload interface{}, description string) {
	// Execute in background to not block main request
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("AuditLog panic recovered", zap.Any("panic", r))
			}
		}()

		var oldJSON, newJSON datatypes.JSON

		if oldPayload != nil {
			b, err := json.Marshal(oldPayload)
			if err == nil {
				oldJSON = datatypes.JSON(b)
			}
		}

		if newPayload != nil {
			b, err := json.Marshal(newPayload)
			if err == nil {
				newJSON = datatypes.JSON(b)
			}
		}

		log := &models.AuditLog{
			EmployeeID:  actorID,
			Action:      action,
			EntityType:  entityType,
			EntityID:    entityID,
			OldValues:   oldJSON,
			NewValues:   newJSON,
			Description: description,
		}

		// TODO: Add IP and UserAgent if passed through context in future

		if err := s.repo.Create(log); err != nil {
			logger.Error("Failed to create audit log", zap.Error(err))
		}
	}()
}

func (s *auditLogService) GetLogs(filter *auditlog.AuditLogFilter) (*auditlog.ListAuditLogResponse, error) {
	logs, total, err := s.repo.FindAll(filter)
	if err != nil {
		logger.Error("Failed to fetch audit logs", zap.Error(err))
		return nil, err
	}

	responses := make([]auditlog.AuditLogResponse, len(logs))
	for i, log := range logs {
		var oldMap, newMap map[string]interface{}

		if len(log.OldValues) > 0 {
			json.Unmarshal(log.OldValues, &oldMap)
		}
		if len(log.NewValues) > 0 {
			json.Unmarshal(log.NewValues, &newMap)
		}

		userName := "System"
		if log.Employee != nil {
			userName = log.Employee.FullName
		}

		responses[i] = auditlog.AuditLogResponse{
			ID:           log.ID,
			EmployeeID:   log.EmployeeID,
			EmployeeName: userName,
			Action:       log.Action,
			EntityType:   log.EntityType,
			EntityID:     log.EntityID,
			OldValues:    oldMap,
			NewValues:    newMap,
			Description:  log.Description,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt,
		}
	}

	return &auditlog.ListAuditLogResponse{
		Logs:  responses,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}

func (s *auditLogService) ExportLogs(filter *auditlog.AuditLogFilter) ([]byte, string, error) {
	filter.IgnorePagination = true
	logs, _, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	headers := []string{"Time", "Actor", "Action", "Entity Type", "Entity ID", "Description", "IP Address", "User Agent", "Old Values", "New Values"}
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	for _, log := range logs {
		actorName := "System"
		if log.Employee != nil {
			actorName = log.Employee.FullName
		}

		oldValStr := ""
		if len(log.OldValues) > 0 {
			oldValStr = string(log.OldValues)
		}

		newValStr := ""
		if len(log.NewValues) > 0 {
			newValStr = string(log.NewValues)
		}

		record := []string{
			log.CreatedAt.Format(time.RFC3339),
			actorName,
			log.Action,
			log.EntityType,
			log.EntityID.String(),
			log.Description,
			log.IPAddress,
			log.UserAgent,
			oldValStr,
			newValStr,
		}

		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}
