# Future Feature: HR Email Templates & Custom Sending

**Goal**: Enable HR to create and send custom email templates to employees directly from the internal system, moving away from hardcoded HTML in the codebase.

## Current State
- Use `pkg/email/brevo_service.go` with hardcoded HTML strings.
- Only supports specific system emails (Welcome, Password Reset, Birthday).

## Technical Strategy
We will implement a **Database-Backed Template System** combined with a **Generic Email Sender**.

### 1. Database Schema (`email_templates`)
Create a new table to store email drafts and templates.

```go
type EmailTemplate struct {
    ID          uuid.UUID `gorm:"primaryKey"`
    Name        string    `gorm:"unique;not null"` // e.g., "Monthly Newsletter"
    Subject     string    `gorm:"not null"`        // e.g., "News for {{month}}"
    Body        string    `gorm:"type:text"`       // HTML content with placeholders
    Description string    `gorm:"type:text"`
    CreatedBy   uuid.UUID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 2. API Endpoints
- **CRUD for Templates**:
  - `POST /api/v1/email-templates` (Create)
  - `GET /api/v1/email-templates` (List)
  - `PUT /api/v1/email-templates/:id` (Update)
  - `DELETE /api/v1/email-templates/:id` (Delete)

- **Sending Endpoint**:
  - `POST /api/v1/email/send`
  - Body:
    ```json
    {
      "template_id": "uuid...",
      "recipients": ["uuid-1", "uuid-2"], // Employee IDs
      "variables": {
        "month": "October",
        "meeting_link": "..."
      }
    }
    ```

### 3. Refactoring `BrevoService`
- **Current**: Detailed methods like `SendWelcomeEmail()`.
- **New Generic Method**:
  ```go
  func (s *brevoEmailService) SendGenericEmail(to []string, subject string, htmlContent string) error {
      // 1. Validate inputs
      // 2. Call Brevo API
      // 3. Log result
  }
  ```

### 4. Frontend Requirements
- **Rich Text Editor**: Use a library like `React-Quill` or `TinyMCE`.
- **Variable Picker**: Allow HR to insert placeholders like `{{first_name}}`, `{{employee_code}}` into the editor.
- **Preview Mode**: Show a rendered email with dummy data before sending.

## Benefits
1.  **Flexibility**: HR can change wording without code deploys.
2.  **Power**: Supports bulk sending to departments or custom lists.
3.  **Clean Code**: Removes massive HTML strings from Go files.
