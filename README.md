# Simple Appointment Booking API

## Overview
Build a basic appointment booking API that handles scheduling across multiple coaches.

## The Challenge

Build a simple scheduling system with 4 endpoints that:
- Shows available appointment slots
- Books appointments without double-booking
- Handles webhook events from external systems
- Shows how appointments are distributed among coaches

### APIs Provided

#### 1. Get Available Slots
`GET /api/availability`
- Return available appointment times for the next 7 days
- Aggregate slots from multiple coaches
- Handle timezones (store in UTC)

#### 2. Book an Appointment
`POST /api/appointments`
- Prevent double-booking (race conditions) with SQL Transactions
- Assign to a coach based on their performance score
- Sends a webhook to the mock CRM service

#### 3. Handle Webhooks
`POST /api/webhooks/calendar`
- Processes updates from external calendar (mocked)
- Handles duplicates (idempotency) - idempotency keys stored in Redis
- Retries failed webhook calls - with exponential backoff + jitter

#### 4. Show Distribution
`GET /api/coaches/distribution`
- Display how appointments are distributed
- Includes basic metrics per coach

#### 5. Health Check Endpoint
`GET /health`


#### Utilization and Fairness Score Calculation

utlization = appointments / maximum_daily_appointments

fairness_score = 1 - standard_deviation(utilization)


## Getting Started (10 minutes)

### Prerequisites
- Docker and Docker Compose
- Go

### Setup Steps

```bash
# 1. Clone and setup
git clone <repo>
cd backend-devops-challenge
cp .env.example .env

# 2. Start services (PostgreSQL, Redis, Mock APIs)
docker-compose up -d


### Database Connection
```
postgresql://postgres:postgres123@localhost:5432/scheduling_db
```

### Make Sure It Runs
```bash
docker-compose up
```

**Testing database connection:**
```bash
docker exec -it scheduling-postgres psql -U postgres -d scheduling_db -c '\dt'
```
