package notification

import (
	"log"
)

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) NotifyPractitioner(practitionerEmail string, sampleID string, patientName string) error {
	log.Printf("[EMAIL STUB] Sent notification to %s: Result for sample %s (Patient: %s) is READY.", 
		practitionerEmail, sampleID, patientName)
	return nil
}