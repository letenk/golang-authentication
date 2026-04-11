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
	m.SetHeader("Subject", "Your Password Reset OTP")
	m.SetBody("text/html", fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Use the OTP code below to reset your password:</p>
		<h2>%s</h2>
		<p>This code is valid for 5 minutes and can only be used once.</p>
		<p>If you did not request a password reset, please ignore this email.</p>
	`, toName, otp))

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	return d.DialAndSend(m)
}
