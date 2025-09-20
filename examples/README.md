# API Examples & Testing

## Quick Start

All examples use API key authentication:
```bash
-H "X-API-Key: test-key-123"
```

## 1. Check Available Slots

```bash
curl "http://localhost:3000/api/availability?days=7" \
  -H "X-API-Key: test-key-123"
```

Expected response:
```json
{
  "slots": [
    {
      "coach_id": "coach-1",
      "start_time": "2024-01-15T14:00:00Z",
      "end_time": "2024-01-15T14:30:00Z"
    }
  ]
}
```

## 2. Book an Appointment

```bash
curl -X POST "http://localhost:3000/api/appointments" \
  -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{
    "calendar_id": "cal-1",
    "contact_email": "john@example.com",
    "contact_name": "John Doe",
    "start_time": "2024-01-15T14:00:00Z"
  }'
```

Success (201):
```json
{
  "appointment_id": "apt-123",
  "coach_id": "coach-1",
  "status": "scheduled"
}
```

Conflict (409):
```json
{
  "error": "SLOT_UNAVAILABLE",
  "message": "This time slot is no longer available"
}
```

## 3. Handle Webhook

```bash
curl -X POST "http://localhost:3000/api/webhooks/calendar" \
  -H "X-Idempotency-Key: unique-key-123" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "appointment.cancelled",
    "appointment_id": "apt-123"
  }'
```

## 4. Check Distribution

```bash
curl "http://localhost:3000/api/coaches/distribution" \
  -H "X-API-Key: test-key-123"
```

## Critical Test Cases

### Test 1: Prevent Double Booking (Race Condition)

Run these commands simultaneously in two terminals:
```bash
# Terminal 1
curl -X POST "http://localhost:3000/api/appointments" \
  -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"calendar_id":"cal-1","contact_email":"user1@example.com","start_time":"2024-01-15T15:00:00Z"}'

# Terminal 2 (run immediately)
curl -X POST "http://localhost:3000/api/appointments" \
  -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"calendar_id":"cal-1","contact_email":"user2@example.com","start_time":"2024-01-15T15:00:00Z"}'
```

**Expected**: One succeeds, one returns 409 conflict

### Test 2: Webhook Idempotency

Send the same webhook twice:
```bash
# First request
curl -X POST "http://localhost:3000/api/webhooks/calendar" \
  -H "X-Idempotency-Key: test-key-001" \
  -H "Content-Type: application/json" \
  -d '{"event_type":"appointment.updated","appointment_id":"apt-123"}'

# Same request again
curl -X POST "http://localhost:3000/api/webhooks/calendar" \
  -H "X-Idempotency-Key: test-key-001" \
  -H "Content-Type: application/json" \
  -d '{"event_type":"appointment.updated","appointment_id":"apt-123"}'
```

**Expected**: Both return success, but only processed once

### Test 3: CRM Webhook Retry

The mock CRM fails 30% of requests. Your implementation should retry:
```bash
# Book appointment and check if webhook was sent
curl -X POST "http://localhost:3000/api/appointments" \
  -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"calendar_id":"cal-1","contact_email":"test@example.com","start_time":"2024-01-15T16:00:00Z"}'

# Check CRM webhook logs
curl http://localhost:3002/webhooks/log
```

**Expected**: Retries on failure with exponential backoff

## Mock Service Endpoints

### Test Mock Services Directly

```bash
# Get coach availability from mock calendar API
curl "http://localhost:3001/coaches/coach-1/availability?days=2"

# Check mock CRM webhook logs
curl http://localhost:3002/webhooks/log

# Validate API key with mock auth service
curl -X POST http://localhost:3003/validate \
  -H "Content-Type: application/json" \
  -d '{"api_key": "test-key-123"}'
```

## Database Verification

```bash
# Check for double bookings (should return 0 rows)
docker exec scheduling-postgres psql -U postgres -d scheduling_db -c \
  "SELECT coach_id, start_time, COUNT(*) FROM coach_appointments 
   WHERE status = 'scheduled' 
   GROUP BY coach_id, start_time 
   HAVING COUNT(*) > 1;"
```