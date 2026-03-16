package email

import (
	"context"
	"fmt"
	"time"

	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/sendinblue/APIv3-go-library/v2/lib"
	"go.uber.org/zap"
)

// EmailService interface for sending emails
type EmailService interface {
	SendWelcomeEmail(to, name, tempPassword string) error
	SendPasswordResetEmail(to, name, resetLink string) error
	SendPasswordChangedEmail(to, name string) error
	SendBirthdayWish(to, name string) error
}

type brevoEmailService struct {
	client    *lib.APIClient
	fromEmail string
	fromName  string
	apiKey    string
}

// NewBrevoEmailService creates a new Brevo email service
func NewBrevoEmailService(apiKey, fromEmail, fromName string) EmailService {
	cfg := lib.NewConfiguration()
	// Set API key in configuration
	cfg.AddDefaultHeader("api-key", apiKey)

	client := lib.NewAPIClient(cfg)

	// Debug logging (masking most of the key)
	keyLen := len(apiKey)
	prefix := ""
	if keyLen > 5 {
		prefix = apiKey[:5]
	}
	logger.Info("Brevo email service initialized",
		zap.String("from_email", fromEmail),
		zap.String("from_name", fromName),
		zap.Int("api_key_len", keyLen),
		zap.String("api_key_prefix", prefix),
	)

	return &brevoEmailService{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
		apiKey:    apiKey,
	}
}

// SendWelcomeEmail sends welcome email with temporary password
func (s *brevoEmailService) SendWelcomeEmail(to, name, tempPassword string) error {
	subject := "Welcome to Internal Collaboration System"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .credentials { background: white; padding: 20px; border-left: 4px solid #667eea; margin: 20px 0; }
        .password { font-size: 24px; font-weight: bold; color: #667eea; letter-spacing: 2px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome Aboard!</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>Welcome to our Internal Collaboration System! Your account has been created by the HR team.</p>
            
            <div class="credentials">
                <h3>Your Login Credentials:</h3>
                <p><strong>Email:</strong> %s</p>
                <p><strong>Temporary Password:</strong></p>
                <p class="password">%s</p>
            </div>
            
            <p><strong>⚠️ Important:</strong> This is a temporary password. You will be required to change it on your first login.</p>
            
            <h3>Next Steps:</h3>
            <ol>
                <li>Visit the login page</li>
                <li>Use the credentials above to log in</li>
                <li>You'll be prompted to set a new password</li>
                <li>Complete your profile setup</li>
            </ol>
            
            <p>If you have any questions, please contact the HR department.</p>
            
            <p>Best regards,<br>
            <strong>HR Team</strong></p>
        </div>
        <div class="footer">
            <p>This is an automated email. Please do not reply to this message.</p>
        </div>
    </div>
</body>
</html>
	`, name, to, tempPassword)

	textContent := fmt.Sprintf(`
Welcome to Internal Collaboration System!

Hello %s,

Your account has been created. Here are your login credentials:

Email: %s
Temporary Password: %s

IMPORTANT: This is a temporary password. You will be required to change it on your first login.

Next Steps:
1. Visit the login page
2. Use the credentials above to log in
3. You'll be prompted to set a new password
4. Complete your profile setup

If you have any questions, please contact the HR department.

Best regards,
HR Team
	`, name, to, tempPassword)

	return s.sendEmail(to, name, subject, htmlContent, textContent)
}

// SendPasswordResetEmail sends password reset email
func (s *brevoEmailService) SendPasswordResetEmail(to, name, resetLink string) error {
	subject := "Password Reset Request"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #f5576c; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Password Reset</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            
            <div style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </div>
            
            <div class="warning">
                <strong>⚠️ Security Notice:</strong>
                <ul>
                    <li>This link will expire in 1 hour</li>
                    <li>If you didn't request this, please ignore this email</li>
                    <li>Never share this link with anyone</li>
                </ul>
            </div>
            
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;">%s</p>
            
            <p>Best regards,<br>
            <strong>Security Team</strong></p>
        </div>
        <div class="footer">
            <p>This is an automated email. Please do not reply to this message.</p>
        </div>
    </div>
</body>
</html>
	`, name, resetLink, resetLink)

	textContent := fmt.Sprintf(`
Password Reset Request

Hello %s,

We received a request to reset your password. Click the link below to create a new password:

%s

Security Notice:
- This link will expire in 1 hour
- If you didn't request this, please ignore this email
- Never share this link with anyone

Best regards,
Security Team
	`, name, resetLink)

	return s.sendEmail(to, name, subject, htmlContent, textContent)
}

// SendPasswordChangedEmail sends notification when password is changed
func (s *brevoEmailService) SendPasswordChangedEmail(to, name string) error {
	subject := "Password Changed Successfully"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #4facfe 0%%, #00f2fe 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .success { background: #d4edda; border-left: 4px solid #28a745; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Password Changed</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            
            <div class="success">
                <strong>✓ Success!</strong> Your password has been changed successfully.
            </div>
            
            <p>Your account password was recently updated. If you made this change, no further action is needed.</p>
            
            <p><strong>⚠️ Didn't make this change?</strong></p>
            <p>If you didn't change your password, please contact the IT security team immediately at <a href="mailto:security@company.com">security@company.com</a></p>
            
            <p>Best regards,<br>
            <strong>Security Team</strong></p>
        </div>
        <div class="footer">
            <p>This is an automated email. Please do not reply to this message.</p>
        </div>
    </div>
</body>
</html>
	`, name)

	textContent := fmt.Sprintf(`
Password Changed Successfully

Hello %s,

Your password has been changed successfully.

Your account password was recently updated. If you made this change, no further action is needed.

Didn't make this change?
If you didn't change your password, please contact the IT security team immediately at security@company.com

Best regards,
Security Team
	`, name)

	return s.sendEmail(to, name, subject, htmlContent, textContent)
}

// SendBirthdayWish sends a happy birthday email
func (s *brevoEmailService) SendBirthdayWish(to, name string) error {
	subject := "🎂 Happy Birthday from the Team!"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 20px auto; padding: 20px; background: #fff; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #FF9A9E 0%%, #FECFEF 100%%); color: white; padding: 40px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { padding: 40px 30px; text-align: center; }
        .message { font-size: 18px; margin: 20px 0; color: #555; }
        .highlight { font-size: 24px; font-weight: bold; color: #FF6B6B; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #999; font-size: 12px; border-top: 1px solid #eee; padding-top: 20px; }
        .emoji { font-size: 48px; margin-bottom: 20px; display: block; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Happy Birthday!</h1>
        </div>
        <div class="content">
            <span class="emoji">🎉🎂🎈</span>
            <h2>Dear %s,</h2>
            
            <p class="message">Wishing you a fantastic birthday filled with joy, laughter, and cake!</p>
            
            <div class="highlight">
                Have an amazing day!
            </div>
            
            <p class="message">We are lucky to have you on our team. Here's to another year of success and happiness.</p>
            
            <p>Best wishes,<br>
            <strong>Your Colleagues</strong></p>
        </div>
        <div class="footer">
            <p>Internal Collaboration System</p>
        </div>
    </div>
</body>
</html>
	`, name)

	textContent := fmt.Sprintf(`
Happy Birthday from the Team!

Dear %s,

Wishing you a fantastic birthday filled with joy, laughter, and cake!
Have an amazing day!

We are lucky to have you on our team. Here's to another year of success and happiness.

Best wishes,
Your Colleagues
	`, name)

	return s.sendEmail(to, name, subject, htmlContent, textContent)
}

// sendEmail is the core method to send email via Brevo
func (s *brevoEmailService) sendEmail(to, toName, subject, htmlContent, textContent string) error {
	// Create a fresh client for each request to avoid concurrency issues
	cfg := lib.NewConfiguration()
	cfg.AddDefaultHeader("api-key", s.apiKey)
	client := lib.NewAPIClient(cfg)

	// Use context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sendSmtpEmail := lib.SendSmtpEmail{
		Sender: &lib.SendSmtpEmailSender{
			Name:  s.fromName,
			Email: s.fromEmail,
		},
		To: []lib.SendSmtpEmailTo{
			{
				Email: to,
				Name:  toName,
			},
		},
		Subject:     subject,
		HtmlContent: htmlContent,
		TextContent: textContent,
	}

	result, resp, err := client.TransactionalEmailsApi.SendTransacEmail(ctx, sendSmtpEmail)
	if err != nil {
		logger.Error("Failed to send email via Brevo",
			zap.Error(err),
			zap.String("to", to),
			zap.String("subject", subject),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully via Brevo",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("message_id", result.MessageId),
		zap.Int("status_code", resp.StatusCode),
	)

	return nil
}
