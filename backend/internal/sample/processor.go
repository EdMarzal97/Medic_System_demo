package sample

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"vitract-assessment/internal/patient"
)

type Processor struct {
	repository  *Repository
	patientRepo *patient.Repository
}

func NewProcessor(repository *Repository, patientRepo *patient.Repository) *Processor {
	return &Processor{
		repository:  repository,
		patientRepo: patientRepo,
	}
}

func (p *Processor) ProcessFile(ctx context.Context, path string, sampleCode string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read file: %v", err)
		_ = p.repository.UpdateSampleStatusToFailed(ctx, sampleCode, errMsg)
		return sampleCode, fmt.Errorf("%s", errMsg)
	}

	var sampleFile SampleFile
	err = json.Unmarshal(data, &sampleFile)
	if err != nil {
		errMsg := fmt.Sprintf("failed to parse json: %v", err)
		_ = p.repository.UpdateSampleStatusToFailed(ctx, sampleCode, errMsg)
		return sampleCode, fmt.Errorf("%s", errMsg)
	}

	if sampleFile.SampleID != "" {
		sampleCode = sampleFile.SampleID
	}

	_, err = p.patientRepo.GetByPatientID(ctx, sampleFile.PatientID)
	if err != nil {
		errMsg := fmt.Sprintf("validation failed: patient %s not found", sampleFile.PatientID)
		_ = p.repository.UpdateSampleStatusToFailed(ctx, sampleCode, errMsg)
		return sampleCode, fmt.Errorf("%s", errMsg)
	}

	err = p.repository.InsertPendingSample(ctx, sampleCode, sampleFile.PatientID, sampleFile.SequencingType, sampleFile.CollectedDate)
	if err != nil {
		return sampleCode, fmt.Errorf("failed to register sample: %w", err)
	}

	var shannonScore float64
	for _, taxon := range sampleFile.Taxa {
		if taxon.Abundance > 0 {
			shannonScore += taxon.Abundance * math.Log(taxon.Abundance)
		}
	}
	if shannonScore != 0 {
		shannonScore = -shannonScore
	}

	err = p.repository.SaveProcessedResults(ctx, sampleCode, shannonScore, sampleFile.Taxa)
	if err != nil {
		return sampleCode, fmt.Errorf("failed to save results: %w", err)
	}

	return sampleCode, nil
}