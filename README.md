# Vitract Microbiome Lab Platform Assessment

A reliable, scalable, and self-contained microbiome lab platform slice designed to ingest patient lab result files asynchronously, calculate microbial diversity scores, persist the structured data, and surface them through a responsive practitioner dashboard.

---

## System Architecture & Design

To understand how data flows through the system—from the initial ingestion of raw JSON/CSV files to the interactive user interface—please refer to the following architectural blueprints:

### High-Level Architecture

![System Architecture](./.github/assets/architecture.jpg)

### Database Schema Model (PostgreSQL)

![Database Diagram](./.github/assets/db_diagram.jpg)

### Asynchronous Processing Pipeline (Worker)

![Pipeline Diagram](./.github/assets/pipeline.jpg)

---

## 🛠️ Tech Stack

- **Backend & Background Worker:** Go (Golang) 1.22
- **Frontend Dashboard:** React (TypeScript / Vite)
- **Database:** PostgreSQL 16
- **Orchestration & Containerization:** Docker & Docker Compose

---

## 🚀 Getting Started (How to Run)

Follow these exact steps to spin up the entire ecosystem on a fresh machine from a clean clone.

### Prerequisites

- Ensure you have **Docker** and **Docker Compose** installed and running on your system.

### Installation & Execution

1.  **Clone the repository:**

    ```bash
    git clone
    cd vitract-assessment
    ```

2.  **Launch the multi-container environment:**
    Run the following command from the root directory (where the `docker-compose.yml` file is located). This command builds the custom Go backend and React frontend images, provisions the PostgreSQL database, mounts the input data volumes, and links everything on an isolated network:

    ```bash
    docker compose up --build
    ```

3.  **Access the Applications:**
    - **Frontend Dashboard:** Open your browser at [http://localhost:3000](http://localhost:3000)
    - **Backend API Base URL:** [http://localhost:8080](http://localhost:8080)
    - **API Health Check:** [http://localhost:8080/health](http://localhost:8080/health)

## Trade-offs, Alternatives & Future Improvements

Given the constraints of a rapid technical assignment, certain strategic architectural choices were deliberately prioritized over others. With more development time, the following enhancements would be made:

### 1. Dedicated Event Streaming (RabbitMQ / Kafka) vs. Internal Worker Threading

- **Current State:** The current architecture uses a concurrent native Go-routine worker loop pattern. While highly performant for low-to-medium volumes, it shares CPU/Memory bottlenecks with the HTTP API layer.
- **Alternative:** In a true enterprise environment, we would introduce a decoupled message broker like **RabbitMQ** or **Apache Kafka**. The API would simply publish a `SampleIngestedEvent`, and an entirely separate consumer microservice would pull messages off the queue. This guarantees absolute fault isolation and effortless horizontal scaling.

### 2. Live WebSocket Triggers vs. Dashboard Polling

- **Current State:** The frontend displays states populated at the time of page render, needing a manual browser refresh or periodic polling intervals to fetch processing status updates.
- **Alternative:** For non-technical founders and practitioners who need real-time data accuracy, we would hook up a server-sent events (SSE) framework or a full **WebSocket** connection. This would allow the backend worker to push status changes (`pending` ➔ `processing` ➔ `done`) directly to the client view instantly without client overhead.

### 3. Tighter Input Validation & Idempotency

- **Current State:** The pipeline assumes the underlying data files generally align structurally with the assessment boundaries.
- **Alternative:** For bulletproof data pipelines, we would implement an upstream validation layer using strong schemas, as well as strict database-level constraints on sample hashes to enforce absolute request idempotency—preventing accidental processing of identical biological records twice.
