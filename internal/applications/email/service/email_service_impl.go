package service

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type SMTPEmailServiceImpl struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPEmailService(host string, port int, username, password, from string) EmailService {
	return &SMTPEmailServiceImpl{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailServiceImpl) SendOTP(toEmail, toName, otp string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Reset Password OTP")
	m.SetBody("text/html", fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Gunakan kode OTP berikut untuk reset password kamu:</p>
		<h2>%s</h2>
		<p>Kode ini berlaku selama 5 menit dan hanya bisa digunakan sekali.</p>
		<p>Jika kamu tidak meminta reset password, abaikan email ini.</p>
	`, toName, otp))

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	return d.DialAndSend(m)
}
