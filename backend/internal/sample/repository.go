package sample

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) UpsertSample(
	ctx context.Context,
	sample Sample,
) error {
	_, err := r.conn.Exec(
		ctx,
		`
		INSERT INTO samples (
			sample_id,
			patient_id,
			sequencing_type,
			collected_date,
			status
		)
		VALUES ($1, $2, $3, $4, $5)

		ON CONFLICT (sample_id)
		DO UPDATE SET
			patient_id = EXCLUDED.patient_id,
			sequencing_type = EXCLUDED.sequencing_type,
			collected_date = EXCLUDED.collected_date,
			status = EXCLUDED.status,
			error_message = NULL,
			updated_at = NOW()
		`,
		sample.SampleID,
		sample.PatientID,
		sample.SequencingType,
		sample.CollectedDate,
		sample.Status,
	)

	return err
}

func (r *Repository) InsertPendingSample(ctx context.Context, sampleID string, patientID string, sequencingType string, collectedDate string) error {
	var internalPatientID int64
	err := r.conn.QueryRow(ctx, `SELECT id FROM patients WHERE patient_id = $1`, patientID).Scan(&internalPatientID)
	if err != nil {
		return fmt.Errorf("patient not found: %w", err)
	}

	_, err = r.conn.Exec(ctx, `
		INSERT INTO samples (sample_id, patient_id, sequencing_type, collected_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', NOW(), NOW())
		ON CONFLICT (sample_id) DO UPDATE SET
			status = 'pending',
			error_message = NULL,
			processed_at = NULL,
			updated_at = NOW()`,
		sampleID, internalPatientID, sequencingType, collectedDate,
	)
	return err
}

func (r *Repository) GetAllSummaries(ctx context.Context) ([]SampleSummary, error) {
	rows, err := r.conn.Query(
		ctx,
		`SELECT s.sample_id, p.patient_id, s.sequencing_type, s.collected_date, s.status, s.error_message, r.diversity_score
         FROM samples s
         JOIN patients p ON s.patient_id = p.id
         LEFT JOIN results r ON r.sample_id = s.id
         ORDER BY s.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []SampleSummary
	for rows.Next() {
		var s SampleSummary
		err := rows.Scan(
			&s.SampleID,
			&s.PatientID,
			&s.SequencingType,
			&s.CollectedDate,
			&s.Status,
			&s.ErrorMessage,
			&s.DiversityScore,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (r *Repository) GetFullResult(ctx context.Context, sampleID string) (*FullResultView, error) {
	var view FullResultView
	var resultID int64

	err := r.conn.QueryRow(
		ctx,
		`SELECT s.sample_id, p.patient_id, r.id, r.diversity_score
		 FROM samples s
		 JOIN patients p ON s.patient_id = p.id
		 JOIN results r ON r.sample_id = s.id
		 WHERE s.sample_id = $1`,
		sampleID,
	).Scan(&view.SampleID, &view.PatientID, &resultID, &view.DiversityScore)
	
	if err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(
		ctx,
		`SELECT taxon_name, abundance FROM taxa_results WHERE result_id = $1`,
		resultID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Taxon
		if err := rows.Scan(&t.Name, &t.Abundance); err != nil {
			return nil, err
		}
		view.Taxa = append(view.Taxa, t)
	}

	return &view, nil
}

func (r *Repository) SaveProcessedResults(ctx context.Context, sampleCode string, shannonScore float64, taxa []Taxon) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var sampleID int64
	err = tx.QueryRow(ctx, "SELECT id FROM samples WHERE sample_id = $1", sampleCode).Scan(&sampleID)
	if err != nil {
		return fmt.Errorf("sample not found: %w", err)
	}

	var resultID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO results (sample_id, diversity_score) 
		VALUES ($1, $2) 
		ON CONFLICT (sample_id) 
		DO UPDATE SET diversity_score = EXCLUDED.diversity_score
		RETURNING id`,
		sampleID, shannonScore,
	).Scan(&resultID)
	if err != nil {
		return fmt.Errorf("failed to upsert results: %w", err)
	}

	for _, taxon := range taxa {
		_, err = tx.Exec(ctx, `
			INSERT INTO taxa_results (result_id, taxon_name, abundance) 
			VALUES ($1, $2, $3)
			ON CONFLICT (result_id, taxon_name) 
			DO UPDATE SET abundance = EXCLUDED.abundance`,
			resultID, taxon.Name, taxon.Abundance,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert taxon %s: %w", taxon.Name, err)
		}
	}

	_, err = tx.Exec(ctx, "UPDATE samples SET status = 'done', updated_at = NOW() WHERE id = $1", sampleID)
	if err != nil {
		return fmt.Errorf("failed to update sample status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) UpdateSampleStatusToFailed(ctx context.Context, sampleID string, errorMessage string) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO samples (
			sample_id, 
			patient_id, 
			sequencing_type, 
			collected_date, 
			status, 
			error_message, 
			created_at, 
			updated_at
		)
		VALUES (
			$1, 
			(SELECT id FROM patients LIMIT 1),
			'UNKNOWN',
			NOW(),
			'failed',
			$2,
			NOW(),
			NOW()
		)
		ON CONFLICT (sample_id) DO UPDATE SET
			status = 'failed',
			error_message = EXCLUDED.error_message,
			updated_at = NOW()`,
		sampleID, errorMessage,
	)
	return err
}