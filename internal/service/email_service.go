package service

import (
	"crypto/tls"
	"fmt"
	"go-banking-service/internal/logger"

	"github.com/go-mail/mail/v2"
)

type EmailService interface {
	SendPaymentNotification(to string, amount float64) error
}

type emailService struct {
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
}

func NewEmailService(smtpHost string, smtpPort int, smtpUser string, smtpPass string) EmailService {
	return &emailService{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		smtpUser: smtpUser,
		smtpPass: smtpPass,
	}
}

func (s *emailService) SendPaymentNotification(to string, amount float64) error {
	content := fmt.Sprintf(`
        <h1>Спасибо за оплату!</h1>
        <p>Сумма: <strong>%.2f RUB</strong></p>
        <small>Это автоматическое уведомление</small>
    `, amount)

	m := mail.NewMessage()
	m.SetHeader("From", s.smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Платеж успешно проведен")
	m.SetBody("text/html", content)

	d := mail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUser, s.smtpPass)
	d.TLSConfig = &tls.Config{
		ServerName:         s.smtpHost,
		InsecureSkipVerify: false,
	}

	if err := d.DialAndSend(m); err != nil {
		logger.Error("SMTP error", "error", err, "to", to)
		return fmt.Errorf("email sending failed")
	}

	logger.Info("Email sent successfully", "to", to, "amount", amount)
	return nil
}
