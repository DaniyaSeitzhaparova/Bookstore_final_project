package config

import (
	"gopkg.in/gomail.v2"
	"log"
)

func SendEmail(to, subject, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "osakbajtomiris@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "osakbajtomiris@gmail.com", "tksgfbogbuadaobm")
	if err := d.DialAndSend(m); err != nil {
		log.Printf("SMTP send error: %v", err)
	} else {
		log.Printf(" Email sent to %s", to)
	}
}
