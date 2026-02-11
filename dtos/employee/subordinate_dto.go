package employee

import "github.com/google/uuid"

type ListSubordinatesResponse struct {
	Manager      SubordinateManagerRaw `json:"manager"`
	Subordinates []SubordinateSummary  `json:"subordinates"`
	Total        int                   `json:"total"`
}

type SubordinateManagerRaw struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Position string    `json:"position"`
}

type SubordinateSummary struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Position     string    `json:"position"`
	Department   string    `json:"department"`
	Status       string    `json:"status"`
	AvatarUrl    string    `json:"avatar_url"`
}
