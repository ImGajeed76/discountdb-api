package email_reader

import (
	"context"
	"discountdb-api/internal/config"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GMailClient struct {
	service *gmail.Service
	email   string
}

// NewGMailClient creates a new Gmail client
func NewGMailClient(cfg *config.Config) (*GMailClient, error) {
	ctx := context.Background()

	service, err := gmail.NewService(ctx, option.WithAPIKey(cfg.GMailAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %v", err)
	}

	return &GMailClient{
		service: service,
		email:   cfg.GMailUser,
	}, nil
}

// ListEmails returns a list of emails with their details
func (c *GMailClient) ListEmails() ([]Email, error) {
	messages, err := c.service.Users.Messages.List(c.email).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %v", err)
	}

	var emails []Email
	for _, msg := range messages.Messages {
		email, err := c.getMessage(msg.Id)
		if err != nil {
			// Log error but continue processing other messages
			fmt.Printf("Error getting message %s: %v\n", msg.Id, err)
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// Email represents a Gmail email message
type Email struct {
	ID      string
	From    string
	Subject string
	Body    string
}

// getMessage fetches a single email message by ID
func (c *GMailClient) getMessage(messageID string) (Email, error) {
	message, err := c.service.Users.Messages.Get(c.email, messageID).Format("full").Do()
	if err != nil {
		return Email{}, fmt.Errorf("failed to get message: %v", err)
	}

	email := Email{ID: messageID}

	// Extract headers
	for _, header := range message.Payload.Headers {
		switch header.Name {
		case "Subject":
			email.Subject = header.Value
		case "From":
			email.From = header.Value
		}
	}

	// Extract body
	if message.Payload.Body != nil && message.Payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(message.Payload.Body.Data)
		if err != nil {
			return Email{}, fmt.Errorf("failed to decode message body: %v", err)
		}
		email.Body = string(data)
	}

	return email, nil
}
