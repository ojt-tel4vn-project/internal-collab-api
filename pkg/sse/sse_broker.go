package sse

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"go.uber.org/zap"
)

type Event struct {
	EmployeeID uuid.UUID
	Data       interface{}
}

type SSEBroker interface {
	ServeHTTP() gin.HandlerFunc
	Broadcast(employeeID uuid.UUID, data interface{})
	Start()
	Stop()
}

type sseBroker struct {
	clients        map[uuid.UUID]map[chan interface{}]bool
	newClients     chan clientWrapper
	closingClients chan clientWrapper
	events         chan Event
	stop           chan struct{}
	jwtService     crypto.JWTService
}

type clientWrapper struct {
	employeeID uuid.UUID
	ch         chan interface{}
}

func NewSSEBroker(jwtService crypto.JWTService) SSEBroker {
	return &sseBroker{
		clients:        make(map[uuid.UUID]map[chan interface{}]bool),
		newClients:     make(chan clientWrapper),
		closingClients: make(chan clientWrapper),
		events:         make(chan Event, 100),
		stop:           make(chan struct{}),
		jwtService:     jwtService,
	}
}

func (b *sseBroker) Start() {
	go func() {
		logger.Info("SSE Broker started")
		for {
			select {
			case cw := <-b.newClients:
				if b.clients[cw.employeeID] == nil {
					b.clients[cw.employeeID] = make(map[chan interface{}]bool)
				}
				b.clients[cw.employeeID][cw.ch] = true
				logger.Info("SSE Client connected", zap.String("employee_id", cw.employeeID.String()))

			case cw := <-b.closingClients:
				if connections, ok := b.clients[cw.employeeID]; ok {
					delete(connections, cw.ch)
					if len(connections) == 0 {
						delete(b.clients, cw.employeeID)
					}
					close(cw.ch)
				}
				logger.Info("SSE Client disconnected", zap.String("employee_id", cw.employeeID.String()))

			case event := <-b.events:
				logger.Debug("Dispatching SSE event", zap.String("employee_id", event.EmployeeID.String()))
				if connections, ok := b.clients[event.EmployeeID]; ok {
					for ch := range connections {
						select {
						case ch <- event.Data:
						default:
							// Channel full
							logger.Warn("Failed to send SSE event, channel full", zap.String("employee_id", event.EmployeeID.String()))
						}
					}
				}
			case <-b.stop:
				logger.Info("SSE Broker stopped")
				return
			}
		}
	}()
}

func (b *sseBroker) Stop() {
	close(b.stop)
}

func (b *sseBroker) Broadcast(employeeID uuid.UUID, data interface{}) {
	b.events <- Event{
		EmployeeID: employeeID,
		Data:       data,
	}
}

// ServeHTTP returns a Gin HandlerFunc for the SSE endpoint
func (b *sseBroker) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// EventSource in browser doesn't send Authorization header easily,
		// so we support checking the query parameter "token"
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		claims, err := b.jwtService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		employeeID := claims.UserID

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Flush()

		// Create channel for this client
		messageChan := make(chan interface{}, 20)
		cw := clientWrapper{
			employeeID: employeeID,
			ch:         messageChan,
		}

		// Register client
		b.newClients <- cw

		// Listen for connection close
		notify := c.Writer.CloseNotify()

		for {
			select {
			case <-notify:
				// Client disconnected
				b.closingClients <- cw
				return
			case msg := <-messageChan:
				// Write data to stream
				jsonData, err := json.Marshal(msg)
				if err != nil {
					logger.Error("Failed to marshal SSE message", zap.Error(err))
					continue
				}
				fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
				c.Writer.Flush()
			}
		}
	}
}
