package patient

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type Importer struct {
	repository *Repository
}

func NewImporter(repository *Repository) *Importer {
	return &Importer{
		repository: repository,
	}
}

func (i *Importer) ImportFromCSV(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	for _, record := range records[1:] {
		dateOfBirth, err := time.Parse("2006-01-02", record[2])
		if err != nil {
			return fmt.Errorf("invalid date: %w", err)
		}

		patient := Patient{
			PatientID:         record[0],
			Name:              record[1],
			DateOfBirth:       dateOfBirth,
			Email:             record[3],
			PractitionerEmail: "practitioner@sample.com",
		}

		if err := i.repository.Create(ctx, patient); err != nil {
			return fmt.Errorf("failed to create patient %s: %w", patient.PatientID, err)
		}
	}

	return nil
}