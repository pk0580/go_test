package sender

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// EmailSender интерфейс для отправки сообщений
type EmailSender interface {
	Send(recipient, content string) error
}

// SMTPSender реализация через SMTP
type SMTPSender struct {
	host string
	port string
}

// NewSMTPSender создает новый экземпляр SMTPSender
func NewSMTPSender() *SMTPSender {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "mailpit"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "1025"
	}
	return &SMTPSender{
		host: host,
		port: port,
	}
}

// Send отправляет письмо через SMTP
func (s *SMTPSender) Send(recipient, content string) error {
	// В данном учебном проекте Mailpit не требует авторизации
	from := "sender@example.com"
	to := []string{recipient}
	
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Notification\r\n"+
		"\r\n"+
		"%s\r\n", recipient, content))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	
	// Использование smtp.SendMail для простоты
	err := smtp.SendMail(addr, nil, from, to, msg)
	if err != nil {
		return fmt.Errorf("ошибка при отправке email: %w", err)
	}

	log.Printf("Email успешно отправлен на %s через %s", recipient, addr)
	return nil
}

// MockSender для тестирования без SMTP
type MockSender struct{}

func (m *MockSender) Send(recipient, content string) error {
	log.Printf("[MOCK] Отправка сообщения на %s: %s", recipient, content)
	return nil
}
