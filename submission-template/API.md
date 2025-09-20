# API Documentation

## Authentication
All endpoints require an API key header:
```
X-API-Key: your-api-key
```

## Endpoints

### 1. Get Available Slots
```http
GET /api/availability?days=7&timezone=America/New_York
```

**Response:**
```json
{
  "slots": [
    {
      "coach_id": "coach-1",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
    {
      "coach_id": "coach-2",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
    {
      "coach_id": "coach-3",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
    {
      "coach_id": "coach-4",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
    {
      "coach_id": "coach-5",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
    {
      "coach_id": "coach-6",
      "start_time": "2025-09-22T09:00:00Z",
      "end_time": "2025-09-22T09:15:00Z"
    },
  ],
  "total_available": 191
}
```

### 2. Book Appointment
```http
POST /api/appointments
```

**Request:**
```json
{
  "calendar_id": "cal-1",
  "contact_email": "john.doe@example.com",
  "contact_name": "John Doe",
  "start_time": "2025-09-22T09:30:00Z",
  "notes": "First consultation"
}
```

**Success Response (201):**
```json
{
  "appointment_id": "d71edd7e-f405-4ddf-8f5b-6faf5ab7953b",
  "coach_id": "coach-5",
  "start_time": "2025-09-22T09:30:00Z",
  "EndTime": "2025-09-22T10:00:00Z",
  "status": "scheduled"
}
```

**Conflict Response (409):**
```json
{
  "error": "SLOT_UNAVAILABLE",
  "message": "This time slot is no longer available"
}
```

### 3. Webhook Handler
```http
POST /api/webhooks/calendar
```

**Request:**
```json
{
  "event_type": "appointment.cancelled",
  "appointment_id": "70378f67-444c-4b98-86d4-00cdc28527d8"
}
```

**Response:**
```json
{
  "received": true,
  "event_id": "ccf3e7ba-3b93-4d6c-8e1d-04eae74a7407"
}
```

### 4. Coach Distribution
```http
GET /api/coaches/distribution
```

**Response:**
```json
{
  "distribution": [
    {
      "coach_id": "coach-1",
      "name": "Alice Johnson",
      "Email": "alice@example.com",
      "Score": 0.85,
      "appointments_count": 4,
      "utilization": 0.4
    },
    {
      "coach_id": "coach-2",
      "name": "Bob Smith",
      "Email": "bob@example.com",
      "Score": 0.72,
      "appointments_count": 0,
      "utilization": 0
    },
    {
      "coach_id": "coach-3",
      "name": "Carol Davis",
      "Email": "carol@example.com",
      "Score": 0.9,
      "appointments_count": 7,
      "utilization": 0.7
    },
    {
      "coach_id": "coach-4",
      "name": "David Wilson",
      "Email": "david@example.com",
      "Score": 0.65,
      "appointments_count": 0,
      "utilization": 0
    },
    {
      "coach_id": "coach-5",
      "name": "Eve Martinez",
      "Email": "eve@example.com",
      "Score": 0.58,
      "appointments_count": 0,
      "utilization": 0
    },
    {
      "coach_id": "coach-6",
      "name": "Frank Thompson",
      "Email": "frank@example.com",
      "Score": 0.78,
      "appointments_count": 1,
      "utilization": 0.1
    }
  ],
  "fairness_score": 0.735424868893541
}
```

## Error Responses

All errors follow this format:
```json
{
  "error": "ERROR_CODE",
  "message": "Human readable message",
  "details": {}
}
```

### Common Error Codes
- `VALIDATION_ERROR` - Invalid input
- `AUTHENTICATION_FAILED` - Invalid API key
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `SLOT_UNAVAILABLE` - Booking conflict
- `INTERNAL_ERROR` - Server error

## Rate Limiting
- Default: 100 requests per minute
- Headers included in response:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
  - `X-RateLimit-Reset`

## Testing the API

### Quick Test Commands
```bash
# Check availability
curl -H "X-API-Key: test-key-123" \
  http://localhost:3000/api/availability

# Book appointment
curl -X POST -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"calendar_id":"cal-1","contact_email":"test@example.com","start_time":"2024-01-15T14:00:00Z"}' \
  http://localhost:3000/api/appointments
```
