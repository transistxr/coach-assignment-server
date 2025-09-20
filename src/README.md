# Source Code Directory

This is where you should implement your solution.

## Getting Started

Choose your preferred language and framework to implement the required API endpoints. The solution should:

1. Connect to PostgreSQL database (connection details in `.env`)
2. Implement the 4 required endpoints
3. Integrate with the mock services
4. Handle concurrent requests safely
5. Include tests for critical functionality

## Implementation Requirements

### Required Endpoints
- `GET /api/availability` - Return available appointment slots
- `POST /api/appointments` - Book an appointment
- `POST /api/webhooks/calendar` - Handle webhook events
- `GET /api/coaches/distribution` - Show appointment distribution

### Key Considerations
- Database transactions for booking operations
- Proper error handling and status codes
- Input validation
- Concurrent request handling
- Retry logic for external service calls

## Testing

Include tests for:
- Race condition prevention in bookings
- Webhook idempotency
- Coach distribution logic
- Error scenarios

## Notes
- Use any project structure that makes sense for your chosen framework
- Feel free to use any libraries or ORMs you're comfortable with
- Focus on correctness and clarity over premature optimization