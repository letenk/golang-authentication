package service

type EmailService interface {
	SendOTP(toEmail, toName, otp string) error
}
