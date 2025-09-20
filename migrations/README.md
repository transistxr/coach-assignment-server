# Database Migrations

## Overview
The `init.sql` file contains the database schema with sample data for testing.

## Setup
Migrations run automatically when you start Docker Compose:
```bash
docker-compose up -d
```

## Schema

### Core Tables
- **coaches**: Coach profiles with performance scores
- **coach_appointments**: Appointment bookings with webhook tracking
- **coach_slots**: Available time slots per coach
- **calendars**: Calendar configurations
- **coach_calendars**: Maps coaches to calendars

### Supporting Tables
- **webhook_events**: Webhook event log
- **distribution_log**: Appointment distribution tracking
- **api_keys**: API authentication

## Key Constraints
- Unique constraint on (coach_id, start_time) prevents double booking
- All timestamps stored as TIMESTAMPTZ in UTC
- Appointment status validated against allowed values

## Sample Data
- 6 coaches with varying performance scores
- 3 calendar types
- 10 coach-calendar mappings
- 3 test API keys

## Helper Functions
- `is_slot_available()`: Check slot availability
- `update_updated_at_column()`: Auto-update timestamps

## Manual Connection
```bash
# Connect to database
docker exec -it scheduling-postgres psql -U postgres -d scheduling_db

# List tables
\dt

# Query data
SELECT * FROM coaches;
```