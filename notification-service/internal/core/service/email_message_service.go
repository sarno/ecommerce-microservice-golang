package service

import (
	"crypto/tls"
	"notification-service/config"

	"github.com/go-mail/mail"
	"github.com/labstack/gommon/log"
)
type IEmailService interface {
	SendEmail(to string, subject string, body string) error
}

// struct

type emailService struct {
	Username string
	Password string
	Host     string
	Port     int
	From     string
	IsTls    bool
}

// SendEmail implements [IEmailMessage].
func (e *emailService) SendEmail(to string, subject string, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)

	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(e.Host, e.Port, e.Username, e.Password)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	err := d.DialAndSend(m)
	if err != nil {
		log.Errorf("[SendEmailNotif-1] Error: %v", err)
		return err
	}
	
	return nil
}

// NewEmailMessage
func NewEmailService(cfg *config.Config) IEmailService {
	return &emailService{
		Username: cfg.EmailConfig.Username,
		Password: cfg.EmailConfig.Password,
		Host:     cfg.EmailConfig.Host,
		Port:     cfg.EmailConfig.Port,
		From:     cfg.EmailConfig.Sending,
		IsTls:    cfg.EmailConfig.IsTLS,
	}
}
