package patient

import "time"

type Patient struct {
	ID                int64
	PatientID         string
	Name              string
	DateOfBirth       time.Time
	Email             string
	PractitionerEmail string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}