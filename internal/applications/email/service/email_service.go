package service

type EmailService interface {
	SendOTP(toEmail, toName, otp string) error
	SendVerificationOTP(toEmail, toName, otp string) error
}
