package sample

import "time"

type Taxon struct {
	Name      string  `json:"name"`
	Abundance float64 `json:"abundance"`
}

type SampleFile struct {
	SampleID       string  `json:"sample_id"`
	PatientID      string  `json:"patient_id"`
	SequencingType string  `json:"sequencing_type"`
	CollectedDate  string  `json:"collected_date"`
	Taxa           []Taxon `json:"taxa"`
}

type Sample struct {
	ID             int64
	SampleID       string
	PatientID      int64
	SequencingType string
	CollectedDate  time.Time
	Status         string
	ErrorMessage   *string
}

type SampleSummary struct {
	SampleID       string    `json:"sample_id"`
	PatientID      string    `json:"patient_id"`
	SequencingType string    `json:"sequencing_type"`
	CollectedDate  time.Time    `json:"collected_date"`
	Status         string    `json:"status"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	DiversityScore *float64  `json:"diversity_score,omitempty"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
}

type FullResultView struct {
	SampleID       string  `json:"sample_id"`
	PatientID      string  `json:"patient_id"`
	DiversityScore float64 `json:"diversity_score"`
	Taxa           []Taxon `json:"taxa"`
}