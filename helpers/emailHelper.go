package helpers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

func GetEmailConfig() *EmailConfig {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     port,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SendOTPEmail(toEmail, otp string) error {
	config := GetEmailConfig()
	if config.SMTPUsername == "" || config.SMTPPassword == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	// Build message
	subject := "Your OTP for Login"
	body := fmt.Sprintf("Your OTP code is: %s\nThis code will expire in 10 minutes.\nIf you didn't request this code, please ignore this email.", otp)
	message := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		toEmail, config.FromEmail, subject, body,
	)

	host := config.SMTPHost
	port := config.SMTPPort
	addr := fmt.Sprintf("%s:%d", host, port)

	// 1) Dial in plain-text
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP: %w", err)
	}
	defer client.Close()

	// 2) Upgrade to TLS
	tlsConfig := &tls.Config{
		ServerName: host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// 3) Authenticate
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to auth: %w", err)
	}

	// 4) Send the mail
	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("failed to set MAIL FROM: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open DATA: %w", err)
	}
	if _, err = w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// 5) Cleanly QUIT
	if err = client.Quit(); err != nil {
		return fmt.Errorf("failed to quit SMTP: %w", err)
	}

	log.Printf("OTP email sent successfully to %s", toEmail)
	return nil
}