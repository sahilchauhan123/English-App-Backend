package smtp

import (
	"net/smtp"
	"os"
)

func sendEmail(to string, subject string, body string) error {

	host := "smtp.gmail.com"
	port := "587"
	password := os.Getenv("SMTP_PASSWORD") // Use your SMTP password here
	from := os.Getenv("FROM")

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte("Subject: "+subject+"\r\n\r\n"+body))
	if err != nil {
		return err
	}
	return nil
}

func SendEmailOTP(to string, otp string) error {
	subject := "Your OTP Code"
	body := "Your OTP code is: " + otp
	err := sendEmail(to, subject, body)
	if err != nil {
		return err
	}
	return nil
}
