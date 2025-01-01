package utils

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"
)

// SendEmail sends an email with the provided subject and HTML body to the list of recipients
func SendEmail(from string, password string, recipients string, subject string, body string) error {
	// Define the SMTP server and authentication details
	smtpServer := "smtp.gmail.com"
	auth := smtp.PlainAuth("", from, password, smtpServer)

	// Create the email message with MIME format
	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)

	// Set the Content-Type for the email to multipart/alternative so it can support both text and HTML
	writer.SetBoundary("boundary")

	// Write the headers (Content-Type, Subject, To)
	msg.Write([]byte(fmt.Sprintf("To: %s\r\n", recipients)))
	msg.Write([]byte(fmt.Sprintf("Subject: %s\r\n", subject)))
	msg.Write([]byte("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n"))

	// Create the HTML part with proper MIME headers
	htmlPartHeader := make(textproto.MIMEHeader)
	htmlPartHeader.Set("Content-Type", "text/html; charset=UTF-8")
	htmlPart, err := writer.CreatePart(htmlPartHeader)
	if err != nil {
		log.Printf("Error creating HTML part: %v", err)
		return fmt.Errorf("failed to create HTML part: %v", err)
	}
	htmlPart.Write([]byte(body)) // Write the HTML body to the part

	// Close the multipart writer (this will finalize the email message)
	writer.Close()

	// Send the email using the SMTP server
	err = smtp.SendMail(smtpServer+":587", auth, from, strings.Split(recipients, ","), msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
