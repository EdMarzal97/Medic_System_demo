package patient

import (
	"context"

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

func (r *Repository) Create(ctx context.Context, patient Patient) error {
	_, err := r.conn.Exec(
		ctx,
		`
		INSERT INTO patients (
			patient_id,
			name,
			date_of_birth,
			email,
			practitioner_email
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (patient_id) 
		DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			practitioner_email = EXCLUDED.practitioner_email,
			updated_at = NOW()
		`,
		patient.PatientID,
		patient.Name,
		patient.DateOfBirth,
		patient.Email,
		patient.PractitionerEmail,
	)

	return err
}

func (r *Repository) GetByPatientID(
	ctx context.Context,
	patientID string,
) (*Patient, error) {
	row := r.conn.QueryRow(
		ctx,
		`
		SELECT
			id,
			patient_id,
			name,
			date_of_birth,
			email,
			practitioner_email,
			created_at,
			updated_at
		FROM patients
		WHERE patient_id = $1
		`,
		patientID,
	)

	var patient Patient

	err := row.Scan(
		&patient.ID,
		&patient.PatientID,
		&patient.Name,
		&patient.DateOfBirth,
		&patient.Email,
		&patient.PractitionerEmail,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &patient, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]Patient, error) {
	rows, err := r.conn.Query(
		ctx,
		`SELECT id, patient_id, name, date_of_birth, email, practitioner_email, created_at, updated_at 
		 FROM patients 
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(
			&p.ID,
			&p.PatientID,
			&p.Name,
			&p.DateOfBirth,
			&p.Email,
			&p.PractitionerEmail,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return patients, nil
}