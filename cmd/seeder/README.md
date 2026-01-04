# CFBD DB Seeder

A Go service that seeds a PostgreSQL database with data from the College Football Data (CFBD) API. The seeder performs automatic schema migration and populates tables in phases to handle foreign key dependencies correctly.

## Overview

The seeder connects to the CFBD API, initializes a PostgreSQL database schema, and populates it with college football data including:

- **Reference Data**: Venues, conferences, play types, stat types, draft teams
- **Teams**: Team information and rosters
- **Games**: Game data, calendars, and scoreboards
- **Stats**: Player and team statistics
- **Recruiting**: Recruit information and rankings
- **Draft**: NFL draft picks and related data

## Prerequisites

- Docker and Docker Compose
- CFBD API key ([Get one here](https://collegefootballdata.com/))

## Quick Start

### 1. Set Your API Key

Edit `docker-compose.yml` and replace `{CFBD_API_KEY}` with your actual API key:

```yaml
environment:
  CFBD_API_KEY: "your-api-key-here"
```

Or set it as an environment variable:

```bash
export CFBD_API_KEY="your-api-key-here"
```

### 2. Start Services

```bash
# Start PostgreSQL and run the seeder
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

The seeder will:
1. Wait for PostgreSQL to be healthy
2. Create the `cfbd` schema if it doesn't exist
3. Run database migrations to create all tables
4. Seed data in phases (currently Phase 1 is implemented)

### 3. Access the Database

#### Using pgAdmin (Web UI)

1. Open http://localhost:5050
2. Login with:
   - Email: `admin@localhost.com`
   - Password: `admin`
3. Register a new server:
   - **Name**: CFBD Local
   - **Host**: `postgres`
   - **Port**: `5432`
   - **Database**: `cfbd`
   - **Username**: `cfbd`
   - **Password**: `cfbd`

#### Using psql (Command Line)

```bash
# Connect via Docker
docker exec -it cfbd-postgres psql -U cfbd -d cfbd

# Or connect from host (port 5433)
psql -h localhost -p 5433 -U cfbd -d cfbd
```

## Configuration

### Environment Variables

| Variable | Description | Default (Docker) |
|----------|-------------|------------------|
| `DATABASE_DSN` | PostgreSQL connection string | `postgres://cfbd:cfbd@postgres:5432/cfbd?sslmode=disable` |
| `CFBD_API_KEY` | CFBD API key (required) | Must be set |

### Database Connection

The default connection details:
- **Host**: `postgres` (Docker service name) or `localhost` (from host)
- **Port**: `5432` (container) or `5433` (host)
- **Database**: `cfbd`
- **Username**: `cfbd`
- **Password**: `cfbd`

## Architecture

### Seeding Phases

The seeder runs in phases to handle foreign key dependencies:

1. **Phase 1**: Global lookups (no dependencies)
   - Conferences
   - Venues
   - Play types
   - Stat types
   - Draft teams

2. **Phase 2**: Teams (depends on venues, conferences)
   - *Not yet implemented*

3. **Phase 3**: Calendars & Games (depends on teams)
   - *Not yet implemented*

4. **Phase 4**: Week/Game Stats (depends on games)
   - *Not yet implemented*

5. **Phase 5**: Season Stats (depends on games)
   - *Not yet implemented*

6. **Phase 6**: Recruiting & Draft (depends on teams)
   - *Not yet implemented*

### Database Schema

All tables are created in the `cfbd` schema. The seeder uses:
- PostgreSQL's `search_path` to automatically find tables in the `cfbd` schema
- GORM for ORM and migrations
- Automatic schema detection to avoid re-initializing existing databases

## Development

### Running Locally (without Docker)

```bash
# Set environment variables
export DATABASE_DSN="postgres://cfbd:cfbd@localhost:5433/cfbd?sslmode=disable"
export CFBD_API_KEY="your-api-key"

# Run the seeder
go run main.go
```

### Building the Docker Image

```bash
docker build -t cfbd-seeder .
```

### Viewing Logs

```bash
# View seeder logs
docker-compose logs seeder

# Follow logs in real-time
docker-compose logs -f seeder

# View PostgreSQL logs
docker-compose logs postgres
```

## Database Management

### Fresh Database

To start with a completely fresh database (removes all data):

```bash
docker-compose down -v
docker-compose up --build
```

The `-v` flag removes the `postgres_data` volume.

### Viewing Tables

```sql
-- List all tables in cfbd schema
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'cfbd' 
ORDER BY table_name;

-- Count tables
SELECT COUNT(*) as table_count
FROM information_schema.tables 
WHERE table_schema = 'cfbd';
```

Or use the one-liner:

```bash
docker exec cfbd-postgres psql -U cfbd -d cfbd -c "\dt cfbd.*"
```

### Database Persistence

Data is persisted in the `postgres_data` Docker volume. To remove it:

```bash
docker-compose down -v
```

## Troubleshooting

### Seeder Fails to Connect

- Ensure PostgreSQL is healthy: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Verify the DSN in `docker-compose.yml`

### API Key Issues

- Verify your API key is set correctly
- Check API key validity at https://collegefootballdata.com/
- Review seeder logs for API errors: `docker-compose logs seeder`

### Schema Migration Errors

- The seeder checks if the database is initialized before running migrations
- If migrations fail, you may need to drop and recreate the database
- Use `docker-compose down -v` to start fresh

### Port Conflicts

If port 5433 is already in use, change it in `docker-compose.yml`:

```yaml
ports:
  - "5434:5432"  # Change 5433 to 5434
```

## Services

The docker-compose setup includes:

- **postgres**: PostgreSQL 16 database
- **seeder**: The CFBD data seeder service
- **pgadmin**: Web-based PostgreSQL administration tool
