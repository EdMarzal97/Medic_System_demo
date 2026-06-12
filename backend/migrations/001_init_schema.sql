CREATE TABLE patients (
    id BIGSERIAL PRIMARY KEY,
    patient_id VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    email VARCHAR(255) NOT NULL,
    practitioner_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE samples (
    id BIGSERIAL PRIMARY KEY,
    sample_id VARCHAR(50) NOT NULL UNIQUE,
    patient_id BIGINT NOT NULL,
    sequencing_type VARCHAR(50) NOT NULL,
    collected_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (
        status IN ('pending', 'processing', 'done', 'failed')
    ),
    error_message TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_samples_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(id)
);

CREATE TABLE results (
    id BIGSERIAL PRIMARY KEY,
    sample_id BIGINT NOT NULL UNIQUE,
    diversity_score DECIMAL(10,6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_results_sample
        FOREIGN KEY (sample_id)
        REFERENCES samples(id)
        ON DELETE CASCADE
);

CREATE TABLE taxa_results (
    id BIGSERIAL PRIMARY KEY,
    result_id BIGINT NOT NULL,
    taxon_name VARCHAR(255) NOT NULL,
    abundance DECIMAL(10,6) NOT NULL CHECK (
        abundance >= 0 AND abundance <= 1
    ),
    CONSTRAINT fk_taxa_results_result
        FOREIGN KEY (result_id)
        REFERENCES results(id)
        ON DELETE CASCADE,
    CONSTRAINT unique_result_taxon 
        UNIQUE (result_id, taxon_name)
);

CREATE INDEX idx_samples_patient_id ON samples(patient_id);
CREATE INDEX idx_samples_status ON samples(status);
CREATE INDEX idx_taxa_results_result_id ON taxa_results(result_id);