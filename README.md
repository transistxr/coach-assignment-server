# Backend Engineering Challenge: Simple Appointment Booking API

## Overview
Build a basic appointment booking API that handles scheduling across multiple coaches. Don't worry about making it perfect - we're looking for a working solution that shows your problem-solving approach.

**Expected Time:** 3-4 hours  
**Submission Deadline:** Within 3 days

## What We're Looking For
A pragmatic solution that:
1. Works correctly for the core requirements
2. Handles basic edge cases (like race conditions)
3. Is reasonably organized and documented
4. Shows your thinking process

## The Challenge

Build a simple scheduling system with 4 endpoints that:
- Shows available appointment slots
- Books appointments without double-booking
- Handles webhook events from external systems
- Shows how appointments are distributed among coaches

### 🔴 Core Requirements (Focus Here First)

#### 1. Get Available Slots
`GET /api/availability`
- Return available appointment times for the next 7 days
- Aggregate slots from multiple coaches
- Handle timezones (store in UTC)

#### 2. Book an Appointment
`POST /api/appointments`
- Prevent double-booking (race conditions)
- Assign to a coach based on their performance score
- Send a webhook to the mock CRM service

#### 3. Handle Webhooks
`POST /api/webhooks/calendar`
- Process updates from external calendar (mocked)
- Handle duplicates (idempotency)
- Retry failed webhook calls

#### 4. Show Distribution
`GET /api/coaches/distribution`
- Display how appointments are distributed
- Include basic metrics per coach

### 🟡 Optional Extras (Only If You Have Time)
- Health check endpoint
- Basic rate limiting
- Handle coach becoming unavailable
- Monitoring/metrics documentation

## Getting Started (10 minutes)

### Prerequisites
- Docker and Docker Compose
- Your preferred language runtime (Node.js, Python, Go, etc.)

### Setup Steps

```bash
# 1. Clone and setup
git clone <repo>
cd backend-devops-challenge
cp .env.example .env

# 2. Start services (PostgreSQL, Redis, Mock APIs)
docker-compose up -d

# 3. Wait a few seconds, then verify everything is running
sleep 5
docker ps | grep -E "postgres|redis|mock"

# 4. Check mock services are healthy
curl http://localhost:3001/health  # Calendar API
curl http://localhost:3002/health  # CRM Webhook
curl http://localhost:3003/health  # Auth Service

# 5. Database is auto-initialized with sample data
# 6. Start coding in /src/
```

### Service Ports
| Service | Port | Description |
|---------|------|-------------|
| Your API | 3000 | Your implementation |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache (optional use) |
| Calendar API | 3001 | Mock service (10% failure rate) |
| CRM Webhook | 3002 | Mock service (30% failure rate) |
| Auth Service | 3003 | API key validation |

### Database Connection
```
postgresql://postgres:postgres123@localhost:5432/scheduling_db
```

## What's Provided

- **PostgreSQL database** with schema and sample data (6 coaches, 3 calendars)
- **Mock services** that simulate external APIs with intentional failures:
  - Calendar API: 10% failure rate, returns coach availability
  - CRM Webhook: 30% failure rate, requires `X-Idempotency-Key` header
  - Auth Service: Validates API keys (`test-key-123`, `dev-key-456`, `prod-key-789`)
- **Example requests** in `/examples/`
- **Environment setup** via Docker Compose

## Technical Notes

### Use Whatever You're Comfortable With
- Any language/framework (Node.js, Python, Go, etc.)
- Any libraries or ORMs you prefer
- Just make sure it connects to PostgreSQL

### Focus On
- Preventing race conditions in bookings
- Basic error handling
- Simple tests for critical paths
- Clear documentation of your approach

### Don't Worry About
- Perfect production code
- Complex monitoring setup
- Extensive test coverage
- Over-engineering the solution

## Submission

### Required (2 documentation files)
1. **Your code** in a GitHub repo
2. **SOLUTION.md** - Explain your approach and any trade-offs
3. **API.md** - Simple documentation of your endpoints

### Optional (if time permits)
- **MONITORING.md** - What metrics you'd track in production
- Additional tests or features

### Make Sure It Runs
```bash
docker-compose up
[your setup commands]
[your start command]
```

## Troubleshooting

**Docker issues:**
```bash
docker-compose down -v  # Reset everything
docker-compose up -d    # Start fresh
```

**PostgreSQL container not showing up:**
```bash
docker ps -a | grep postgres  # Check if it exited
docker logs scheduling-postgres  # Check error logs
```

**Port conflicts:** Edit `.env` to change ports if needed

**Testing database connection:**
```bash
docker exec -it scheduling-postgres psql -U postgres -d scheduling_db -c '\dt'
```

## Evaluation Focus

We'll mainly look at:
- Does it work? (Can we book appointments without issues?)
- How did you prevent race conditions?
- Is the code reasonably organized?
- Can we understand your approach from the documentation?

## Tips

- Start with the core requirements
- Use transactions for booking operations
- Test concurrent bookings manually (run two curl commands simultaneously)
- Document your assumptions
- Check `/examples/README.md` for API examples and test cases

## Questions?

Email `kevinvegda@committedcoaches.com` if anything is unclear.

Remember: We're looking for pragmatic solutions, not perfection. Focus on getting the core functionality working first!