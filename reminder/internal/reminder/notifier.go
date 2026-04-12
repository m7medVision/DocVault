// Package reminder provides email notification delivery.
package reminder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

// EmailConfig holds email provider configuration.
type EmailConfig struct {
	Provider   string
	APIKey     string
	FromEmail  string
	FromName   string
	WebhookURL string // For delivery status updates
}

// ResendEmailProvider implements email sending via Resend.
type ResendEmailProvider struct {
	config EmailConfig
	client *http.Client
	logger *slog.Logger
}

// NewResendEmailProvider creates a new Resend email provider.
func NewResendEmailProvider(config EmailConfig, logger *slog.Logger) *ResendEmailProvider {
	return &ResendEmailProvider{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// EmailData holds data for rendering email templates.
type EmailData struct {
	RecipientName string
	DocumentType  string
	DocumentTitle string
	Issuer        string
	TriggerDate   string
	DaysUntil     int
	ActionURL     string
	Subject       string
}

// ReminderEmail represents a reminder email to be sent.
type ReminderEmail struct {
	To      string
	Subject string
	HTML    string
}

// Send sends an email using the Resend provider.
func (p *ResendEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	payload := map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail),
		"to":      to,
		"subject": subject,
		"html":    body,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	p.logger.Info("email sent successfully",
		"to", to,
		"subject", subject,
	)

	return nil
}

// SendReminderEmail sends a reminder email.
func (p *ResendEmailProvider) SendReminderEmail(ctx context.Context, email *ReminderEmail) error {
	// Resend API payload
	payload := map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail),
		"to":      email.To,
		"subject": email.Subject,
		"html":    email.HTML,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	p.logger.Info("email sent successfully",
		"to", email.To,
		"subject", email.Subject,
	)

	return nil
}

// RenderReminderTemplate renders an HTML reminder email from template.
func RenderReminderTemplate(data *EmailData) (string, error) {
	// HTML template for reminder email (simplified version)
	// In production, use React Email for proper Arabic RTL support
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; direction: %s; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4A90D9; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background: #4A90D9; color: white; text-decoration: none; border-radius: 4px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DocVault Reminder</h1>
        </div>
        <div class="content">
            <p>Hello {{.RecipientName}},</p>
            <p>This is a reminder about an upcoming %s document:</p>
            <ul>
                <li><strong>Type:</strong> %s</li>
                <li><strong>Title:</strong> %s</li>
                <li><strong>Issuer:</strong> %s</li>
                <li><strong>Date:</strong> %s (%d days)</li>
            </ul>
            <p style="text-align: center;">
                <a href="%s" class="button">View Document</a>
            </p>
        </div>
        <div class="footer">
            <p>This email was sent by DocVault. To manage your reminders, visit your account settings.</p>
        </div>
    </div>
</body>
</html>
`,
		// Language direction
		"en", "ltr",
		// Document info (placeholder for template execution)
		data.DocumentType, data.DocumentType, data.DocumentTitle,
		data.Issuer, data.TriggerDate, data.DaysUntil,
		data.ActionURL,
	)

	tmpl, err := template.New("reminder").Parse(html)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ReminderDB interface for reminder database operations.
type ReminderDB interface {
	CreateReminderEvent(ctx context.Context, event *ReminderEvent) error
	UpdateEventStatus(ctx context.Context, eventID, status, errorMsg string) error
	MarkEventSent(ctx context.Context, eventID string, sentAt *time.Time) error
}

// NotificationService handles sending notifications.
type NotificationService struct {
	emailSender EmailSender
	db          ReminderDB
	logger      *slog.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(
	emailSender EmailSender,
	db ReminderDB,
	logger *slog.Logger,
) *NotificationService {
	return &NotificationService{
		emailSender: emailSender,
		db:          db,
		logger:      logger,
	}
}

// SendReminderNotification sends a reminder notification for an event.
func (s *NotificationService) SendReminderNotification(
	ctx context.Context,
	eventID, ruleID, recipientEmail string,
	emailData *EmailData,
) error {
	html, err := RenderReminderTemplate(emailData)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if err := s.emailSender.Send(ctx, recipientEmail, emailData.Subject, html); err != nil {
		s.db.UpdateEventStatus(ctx, eventID, EventStatusFailed, err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}

	now := time.Now()
	s.db.MarkEventSent(ctx, eventID, &now)

	s.logger.Info("reminder notification sent",
		"event_id", eventID,
		"rule_id", ruleID,
		"to", recipientEmail,
	)

	return nil
}
