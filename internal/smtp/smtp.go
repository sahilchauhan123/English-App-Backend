// package smtp

// import (
// 	"net/smtp"
// 	"os"
// )

// func sendEmail(to string, subject string, body string) error {

// 	host := "smtp.gmail.com"
// 	port := "587"
// 	password := os.Getenv("SMTP_PASSWORD") // Use your SMTP password here
// 	from := os.Getenv("FROM")

// 	auth := smtp.PlainAuth("", from, password, host)
// 	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte("Subject: "+subject+"\r\n\r\n"+body))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func SendEmailOTP(to string, otp string) error {
// 	subject := "Your OTP Code"
// 	body := "Your OTP code is: " + otp
// 	err := sendEmail(to, subject, body)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

package smtp

import (
	"fmt"
	"net/smtp"
	"os"
)

func sendEmail(to, subject, body string) error {
	host := "smtp.gmail.com"
	port := "587"
	password := os.Getenv("SMTP_PASSWORD") // Use your SMTP password here
	from := os.Getenv("FROM")

	// Create the MIME header and HTML body
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := []byte("Subject: " + subject + "\r\n" + mime + body)

	// Authentication
	auth := smtp.PlainAuth("", from, password, host)

	// Send the email
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, message)
	if err != nil {
		return err
	}
	return nil
}

func SendEmailOTP(to, otp string) error {
	subject := "Your OTP Code"

	// HTML-styled email body
	body := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Your OTP Code</title>
	</head>
	<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background-color: #f4f4f4; padding: 20px; border-radius: 8px;">
			<h2 style="color: #2c3e50; text-align: center;">Your OTP Code</h2>
			<p>Dear User,</p>
			<p>Your One-Time Password (OTP) is:</p>
			<p style="font-size: 24px; font-weight: bold; color: #e74c3c; text-align: center; background-color: #fff; padding: 10px; border: 1px solid #ddd; border-radius: 4px;">
				` + otp + `
			</p>
			<p>This OTP is valid for the next 5 minutes. Please do not share it with anyone.</p>
			<p>If you didn’t request this, please ignore this email.</p>
			<p style="text-align: center; margin-top: 20px;">
				<a href="https://yourwebsite.com" style="color: #3498db; text-decoration: none;">Visit our website</a>
			</p>
			<p>Best regards,<br>Sahil</p>
		</div>
	</body>
	</html>
	`

	err := sendEmail(to, subject, body)
	if err != nil {
		return err
	}
	fmt.Println("Email sent successfully to", to)
	return nil
}
