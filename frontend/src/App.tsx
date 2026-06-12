/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from "react";

const API_BASE = "http://localhost:8080";

interface Taxon {
  taxon_name: string;
  abundance: number | string;
}

interface FullResult {
  sample_id: string;
  diversity_score: number | string;
  taxa: Taxon[];
}

interface Sample {
  id?: number | string;
  sample_id: string;
  patient_id: string;
  sequencing_type?: string;
  status: "pending" | "processing" | "done" | "failed";
  error_message?: string;
}

interface Patient {
  id: number | string;
  patient_id: string;
  name: string;
}

const styles = {
  container: {
    fontFamily: "sans-serif",
    padding: "20px",
    maxWidth: "1200px",
    margin: "0 auto",
  },
  layout: { display: "flex", gap: "40px" },
  section: { flex: 1 },
  table: {
    width: "100%",
    borderCollapse: "collapse" as const,
    marginTop: "10px",
  },
  th: {
    borderBottom: "2px solid #ddd",
    padding: "10px",
    textAlign: "left" as const,
    backgroundColor: "#f5f5f5",
  },
  td: { borderBottom: "1px solid #ddd", padding: "10px" },
  badge: (status: Sample["status"]) => {
    const colors = {
      done: "#e6f4ea",
      failed: "#fce8e6",
      pending: "#ffe0b2",
      processing: "#e8f0fe",
    };
    const textColors = {
      done: "#137333",
      failed: "#c5221f",
      pending: "#b06000",
      processing: "#1a73e8",
    };
    return {
      backgroundColor: colors[status] || "#eee",
      color: textColors[status] || "#333",
      padding: "4px 8px",
      borderRadius: "4px",
      fontSize: "0.85em",
      fontWeight: "bold",
    };
  },
  card: {
    border: "1px solid #ddd",
    borderRadius: "8px",
    padding: "20px",
    backgroundColor: "#fafafa",
    position: "sticky" as const,
    top: "20px",
  },
};

export default function App() {
  const [samples, setSamples] = useState<Sample[]>([]);
  const [patients, setPatients] = useState<Patient[]>([]);
  const [selectedResult, setSelectedResult] = useState<FullResult | null>(null);
  const [loadingResult, setLoadingResult] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchInitialData = async () => {
      try {
        const [samplesRes, patientsRes] = await Promise.all([
          fetch(`${API_BASE}/samples`).catch(() => null),
          fetch(`${API_BASE}/patients`).catch(() => null),
        ]);

        if (!samplesRes || !patientsRes) {
          setError(
            "Could not connect to the Backend API. Make sure it's running.",
          );
          return;
        }

        let samplesData: any = [];
        let patientsData: any = [];

        try {
          samplesData = await samplesRes.json();
        } catch {
          samplesData = [];
        }
        try {
          patientsData = await patientsRes.json();
        } catch {
          patientsData = [];
        }

        const rawSamples = Array.isArray(samplesData)
          ? samplesData
          : samplesData?.samples || [];
        const normalizedSamples = rawSamples.map((s: any) => ({
          id: s.id || s.ID,
          sample_id: s.sample_id || s.SampleID || s.sampleId || "",
          patient_id: s.patient_id || s.PatientID || s.patientId || "",
          sequencing_type:
            s.sequencing_type || s.SequencingType || s.sequencingType || "N/A",
          status: String(
            s.status || s.Status || "pending",
          ).toLowerCase() as Sample["status"],
          error_message:
            s.error_message || s.ErrorMessage || s.errorMessage || "",
        }));

        const rawPatients = Array.isArray(patientsData)
          ? patientsData
          : patientsData?.patients || [];
        const normalizedPatients = rawPatients.map((p: any) => ({
          id: p.id || p.ID,
          patient_id: p.patient_id || p.PatientID || p.patientId || "",
          name: p.name || p.Name || "Unknown",
        }));

        setSamples(normalizedSamples);
        setPatients(normalizedPatients);

        if (
          (!samplesRes.ok || !patientsRes.ok) &&
          normalizedSamples.length === 0
        ) {
          setError(
            "Backend API responded with an empty dataset or error status.",
          );
        } else {
          setError(null);
        }
      } catch {
        setError("An unexpected frontend parsing error occurred.");
      }
    };

    fetchInitialData();
  }, []);

  const getPatientName = (patientId: string) => {
    const p = patients.find(
      (p) =>
        String(p.patient_id) === String(patientId) ||
        String(p.id) === String(patientId),
    );
    return p ? p.name : "Unknown Patient";
  };

  const handleViewResult = (sampleId: string) => {
    setLoadingResult(true);
    setSelectedResult(null);

    fetch(`${API_BASE}/samples/${sampleId}/result`)
      .then((res) => res.json())
      .then((data: any) => {
        const normalized: FullResult = {
          sample_id:
            data.sample_id || data.SampleID || data.sampleId || sampleId,
          diversity_score:
            data.diversity_score !== undefined
              ? data.diversity_score
              : data.DiversityScore !== undefined
                ? data.DiversityScore
                : 0,
          taxa: (data.taxa || data.Taxa || []).map((t: any) => ({
            taxon_name:
              t.taxon_name ||
              t.TaxonName ||
              t.taxonName ||
              t.name ||
              t.Name ||
              "Unknown",
            abundance:
              t.abundance !== undefined
                ? t.abundance
                : t.Abundance !== undefined
                  ? t.Abundance
                  : 0,
          })),
        };
        setSelectedResult(normalized);
        setLoadingResult(false);
      })
      .catch(() => {
        alert("Result not ready or format mismatch.");
        setLoadingResult(false);
      });
  };

  return (
    <div style={styles.container}>
      <header>
        <h1>Vitract Dashboard</h1>
        {error && <p style={{ color: "red", fontWeight: "bold" }}>{error}</p>}
      </header>

      <hr />

      <div style={styles.layout}>
        <div style={styles.section}>
          <h2>Samples & Patient Status</h2>
          <table style={styles.table}>
            <thead>
              <tr>
                <th style={styles.th}>Sample ID</th>
                <th style={styles.th}>Patient Name</th>
                <th style={styles.th}>Sequencing</th>
                <th style={styles.th}>Status</th>
                <th style={styles.th}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {samples.map((sample) => (
                <tr key={sample.id || sample.sample_id}>
                  <td style={styles.td}>
                    <strong>{sample.sample_id}</strong>
                  </td>
                  <td style={styles.td}>{getPatientName(sample.patient_id)}</td>
                  <td style={styles.td}>{sample.sequencing_type || "N/A"}</td>
                  <td style={styles.td}>
                    <span style={styles.badge(sample.status)}>
                      {sample.status.toUpperCase()}
                    </span>
                    {sample.status === "failed" && sample.error_message && (
                      <div
                        style={{
                          color: "#c5221f",
                          fontSize: "0.8em",
                          marginTop: "4px",
                          maxWidth: "200px",
                        }}
                      >
                        {sample.error_message}
                      </div>
                    )}
                  </td>
                  <td style={styles.td}>
                    {sample.status === "done" && (
                      <button
                        onClick={() => handleViewResult(sample.sample_id)}
                        style={{
                          cursor: "pointer",
                          padding: "6px 12px",
                          borderRadius: "4px",
                          border: "1px solid #ccc",
                        }}
                      >
                        View Analysis
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {samples.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    style={{ ...styles.td, textAlign: "center", color: "#666" }}
                  >
                    No samples processed yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div style={{ ...styles.section, maxWidth: "450px" }}>
          <h2>Result Breakdown</h2>

          {loadingResult && (
            <p>Loading analytical breakdown from database...</p>
          )}

          {!selectedResult && !loadingResult && (
            <div style={{ ...styles.card, color: "#666", textAlign: "center" }}>
              Select a processed sample with status "DONE" to inspect its
              Shannon diversity index and taxon abundances.
            </div>
          )}

          {selectedResult && (
            <div style={styles.card}>
              <h3>Sample: {selectedResult.sample_id}</h3>
              <div
                style={{
                  backgroundColor: "#fff",
                  padding: "15px",
                  borderRadius: "6px",
                  border: "1px solid #eee",
                  marginBottom: "20px",
                }}
              >
                <span
                  style={{ fontSize: "0.9em", color: "#666", display: "block" }}
                >
                  Shannon Diversity Index (H)
                </span>
                <strong style={{ fontSize: "2em", color: "#1a73e8" }}>
                  {Number(selectedResult.diversity_score).toFixed(4)}
                </strong>
              </div>

              <h4>Per-Taxon Breakdown</h4>
              <table style={styles.table}>
                <thead>
                  <tr>
                    <th style={styles.th}>Taxon Name</th>
                    <th style={styles.th}>Relative Abundance</th>
                  </tr>
                </thead>
                <tbody>
                  {selectedResult.taxa &&
                    selectedResult.taxa.map((taxon, idx) => (
                      <tr key={idx}>
                        <td style={styles.td}>
                          <i>{taxon.taxon_name}</i>
                        </td>
                        <td style={styles.td}>
                          {(Number(taxon.abundance) * 100).toFixed(2)}%
                        </td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
