# 📝 GoToDo

A full-stack Todo application built with **Go**, **React (TypeScript)**, **PostgreSQL**, and **OpenSearch** — featuring a **CQRS (Command Query Responsibility Segregation)** architecture for high-performance reads at scale.

> Designed to handle **10M+ records** with fast full-text search, faceted filtering, and paginated results.

### 🌐 Live Demo: [http://kanban.zlv.my.id](http://kanban.zlv.my.id)

---

## ✨ Features

- **CRUD Operations** — Create, read, update, and delete todos
- **6 Status Workflow** — Open → In Progress → Review → Done → On Hold → Canceled
- **Full-Text Search** — Powered by PostgreSQL trigram indexes (standard mode) and OpenSearch (CQRS mode)
- **Faceted Filtering** — Filter by status with real-time count badges
- **Server-Side Pagination** — Efficient pagination for large datasets
- **CQRS Toggle** — Switch between standard PostgreSQL queries and OpenSearch-backed reads at runtime
- **Bulk Data Sync** — CLI tool to synchronize PostgreSQL data to OpenSearch using bulk indexing
- **Passkey Protection** — Mutation endpoints (create, update, delete) require a passkey to prevent unauthorized modifications
- **Dockerized** — Full-stack deployment with a single `docker-compose` command

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        React Frontend                           │
│                    (Vite + TypeScript + TailwindCSS)             │
│              React Query · Axios · Debounced Search             │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP REST API
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go Backend (net/http)                    │
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │   Delivery   │───▶│   Use Case   │───▶│   Repository     │  │
│  │  (HTTP Layer)│    │(Business Logic)   │  (Data Access)   │  │
│  └──────────────┘    └──────────────┘    └────────┬─────────┘  │
│                                                    │            │
│                           ┌────────────────────────┼──────┐    │
│                           ▼                        ▼      │    │
│                    ┌─────────────┐          ┌───────────┐ │    │
│                    │ PostgreSQL  │          │ OpenSearch │ │    │
│                    │  (Command   │          │  (Query   │ │    │
│                    │   + Query)  │          │   Side)   │ │    │
│                    └─────────────┘          └───────────┘ │    │
│                           CQRS Pattern ──────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### CQRS Pattern

| Path                     | Mode               | Description                                       |
| ------------------------ | ------------------ | ------------------------------------------------- |
| `POST/PUT/DELETE /todos` | **Command**        | All writes go to PostgreSQL                       |
| `GET /todos`             | **Standard Query** | Reads from PostgreSQL with trigram search         |
| `GET /cqrs/todos`        | **CQRS Query**     | Reads from OpenSearch for faster full-text search |

---

## 🛠️ Tech Stack

| Layer                | Technology                                                  |
| -------------------- | ----------------------------------------------------------- |
| **Frontend**         | React 18, TypeScript, Vite, TailwindCSS, React Query, Axios |
| **Backend**          | Go 1.26, `net/http`, Clean Architecture                     |
| **Database**         | PostgreSQL 16 with `pg_trgm` extension                      |
| **Search Engine**    | OpenSearch 2.x                                              |
| **Containerization** | Docker, Docker Compose                                      |
| **Dev Tools**        | Air (Go hot-reload), Vite HMR                               |

---

## 📁 Project Structure

```
GoToDo/
├── go/                          # Go backend
│   ├── cmd/
│   │   ├── server/main.go       # HTTP server entrypoint
│   │   └── sync/main.go         # PostgreSQL → OpenSearch sync CLI
│   ├── internal/
│   │   ├── domain/              # Entities, interfaces (Todo, Repository, UseCase)
│   │   ├── delivery/http/       # HTTP handlers (REST endpoints)
│   │   ├── repository/
│   │   │   ├── postgres/        # PostgreSQL implementation
│   │   │   └── opensearch/      # OpenSearch query implementation
│   │   └── usecase/             # Business logic layer
│   ├── structure.sql            # Database schema & indexes
│   ├── migrate.sql              # Schema migration scripts
│   ├── Dockerfile               # Multi-stage Go build
│   └── .air.toml                # Hot-reload configuration
│
├── react/                       # React frontend
│   ├── src/
│   │   ├── api/todoApi.ts       # API client with CQRS toggle
│   │   ├── components/          # TodoList, TodoItem, TodoForm
│   │   └── hooks/               # useTodos, useDebounce
│   ├── nginx.conf               # Production Nginx config
│   ├── Dockerfile               # Multi-stage React build
│   └── package.json
│
├── docker-compose.prod.yml      # Full-stack production deployment
├── docker-compose.yml           # OpenSearch (dev standalone)
├── .env.example                 # Environment variable reference
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [PostgreSQL 16+](https://www.postgresql.org/download/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) (for containerized deployment)

### Option 1: Local Development

**1. Clone the repository**

```bash
git clone https://github.com/rezafahlevi/GoToDo.git
cd GoToDo
```

**2. Set up PostgreSQL**

Create the database and initialize the schema:

```bash
createdb gotodo
psql -d gotodo -f go/structure.sql
```

**3. Start OpenSearch**

```bash
docker-compose up -d
```

**4. Configure the Go backend**

```bash
cd go
cp ../.env.example .env
# Edit .env with your database credentials
```

**5. Run the Go backend**

```bash
# With hot-reload (recommended for development)
air

# Or without hot-reload
go run ./cmd/server
```

**6. Run the React frontend**

```bash
cd react
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173` and the API at `http://localhost:8080`.

**7. (Optional) Sync data to OpenSearch**

If you have data in PostgreSQL and want to enable CQRS mode:

```bash
cd go
go run ./cmd/sync
```

---

### Option 2: Docker Compose (Production)

Run the entire stack with a single command:

```bash
# Copy and configure environment variables
cp .env.example .env

# Build and start all services
docker-compose -f docker-compose.prod.yml up --build -d
```

| Service     | URL                   |
| ----------- | --------------------- |
| Frontend    | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| OpenSearch  | http://localhost:9200 |
| PostgreSQL  | localhost:5432        |

To stop all services:

```bash
docker-compose -f docker-compose.prod.yml down
```

To stop and remove all data volumes:

```bash
docker-compose -f docker-compose.prod.yml down -v
```

---

### 📈 Generating 10M Test Records

To see the true power of the CQRS architecture and OpenSearch, you can generate 10 million random todo records.

**1. Generate the SQL Data**

Run the Python script to generate the massive `insert_10m_lorem_todos.sql` file:

```bash
cd go
python3 dummy.py
```

**2. Load the Data into PostgreSQL**

If running locally:

```bash
psql -d gotodo -f insert_10m_lorem_todos.sql
```

If running via Docker Compose:

```bash
cat insert_10m_lorem_todos.sql | docker exec -i gotodo-postgres psql -U postgres -d gotodo
```

**3. Sync to OpenSearch**

Once the data is in PostgreSQL, run the sync job to bulk-index the 10 million rows into OpenSearch:

```bash
docker exec -it gotodo-backend go run ./cmd/sync
# Or locally: go run ./cmd/sync
```

---

## ⚙️ Environment Variables

| Variable            | Default                                              | Description                                            |
| ------------------- | ---------------------------------------------------- | ------------------------------------------------------ |
| `DATABASE_URL`      | `postgres://postgres:postgres@localhost:5432/gotodo` | PostgreSQL connection string                           |
| `OPENSEARCH_URL`    | `http://localhost:9200`                              | OpenSearch endpoint                                    |
| `PORT`              | `8080`                                               | Go server port                                         |
| `ALLOWED_ORIGIN`    | `*`                                                  | CORS allowed origin (set to your domain in production) |
| `VITE_API_URL`      | `http://localhost:8080`                              | Backend API URL (React build-time)                     |
| `POSTGRES_USER`     | `postgres`                                           | PostgreSQL user (Docker)                               |
| `POSTGRES_PASSWORD` | `postgres`                                           | PostgreSQL password (Docker)                           |
| `POSTGRES_DB`       | `gotodo`                                             | PostgreSQL database name (Docker)                      |
| `APP_PASSKEY`       | *(empty — disabled)*                                 | Passkey for create/update/delete endpoints              |

### 🔒 Passkey Protection

The `APP_PASSKEY` variable controls write-access to the API. When set, all mutation endpoints (`POST`, `PUT`, `DELETE`) require the correct passkey via the `X-Passkey` HTTP header. The frontend will prompt users with a dialog before any write operation.

- **To disable** (local development): Set `APP_PASSKEY=` (empty) — no prompt will appear, all writes are allowed.
- **To enable** (production): Set `APP_PASSKEY=your_secret` — only users who know the passkey can create, update, or delete records.

---

## 📡 API Endpoints

### Standard CRUD (`/todos`)

| Method   | Endpoint                              | Description                                |
| -------- | ------------------------------------- | ------------------------------------------ |
| `GET`    | `/todos?search=&status=&page=&limit=` | List todos with search, filter, pagination |
| `GET`    | `/todos/facets?search=`               | Get status facet counts                    |
| `GET`    | `/todos/{id}`                         | Get a single todo                          |
| `POST`   | `/todos`                              | Create a new todo                          |
| `PUT`    | `/todos/{id}`                         | Update a todo                              |
| `DELETE` | `/todos/{id}`                         | Delete a todo                              |

### CQRS Read-Only (`/cqrs/todos`)

| Method | Endpoint                                   | Description               |
| ------ | ------------------------------------------ | ------------------------- |
| `GET`  | `/cqrs/todos?search=&status=&page=&limit=` | List todos via OpenSearch |
| `GET`  | `/cqrs/todos/facets?search=`               | Get facets via OpenSearch |

---

## 🧪 Database Schema

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS todos (
    id     SERIAL PRIMARY KEY,
    title  VARCHAR(255) NOT NULL,
    status INT NOT NULL DEFAULT 1
);

-- Trigram index for fast LIKE/ILIKE search
CREATE INDEX IF NOT EXISTS trgm_idx_todos_title ON todos USING gin (title gin_trgm_ops);

-- Composite index for status filtering + ordered pagination
CREATE INDEX IF NOT EXISTS idx_todos_status_id ON todos (status ASC, id DESC);
```

### Status Codes

| Code | Status      |
| ---- | ----------- |
| 1    | Open        |
| 2    | In Progress |
| 3    | Review      |
| 4    | Done        |
| 5    | On Hold     |
| 6    | Canceled    |

---

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to open an issue or submit a pull request.

---

<p align="center">
  Built with ❤️ using Go + React
</p>
