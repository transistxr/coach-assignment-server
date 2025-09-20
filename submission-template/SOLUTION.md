# Solution Documentation

## Part 1: Architecture & Design Decisions

### Technology Stack
- **Language**: Go
- **Framework**: Chi
- **Database**: PostgreSQL
- **Additional Tools**: Redis

### Key Design Decisions

#### Appointment Booking Flow

On api/availability, external calendar data is normalized into 15-minute slots and stored in the coach_slots table.

To prevent double-booking and race conditions at booking time:

1. Fetch eligible coaches for the calendar ID.
2. Filter by maximum daily appointments and availability.
3. Pick the coach with the highest performance score.

The entire booking process runs inside a PostgreSQL transaction to ensure consistency under concurrent requests.

#### Coach Distribution Algorithm

1. Very rudimentarily selecting the coach with the highest performance score.

#### Webhook Processing

- Outbound webhooks (to CRM, Auth, etc.) are retried with exponential backoff and jitter for resilience.

Idempotency:

- Downstream idempotency keys are forwarded to prevent duplicate side effects.

- Locally, idempotency keys and response data are stored in Redis.

#### Concurrency Control

PostgreSQL transactions are used to lock the booking process and prevent multiple clients from booking the same slot simultaneously. 

## Part 2: Trade-offs & Constraints

### What I Built vs Ideal Solution

| Feature | What I Built | Ideal Solution | Why the Trade-off |
|---------|--------------|----------------|-------------------|
| Coach Distribution | Pick highest score | Fairness model that is a feedback loop | Lack of algorithmic knowledge |
| Retry Logic | Exponential backoff with jitter | Possibly queues? | Queue infra would be difficult to manage |
| Rate Limiting | 100 requests/minute | IP/User-based + Sliding Window | Scope |

### Time Constraints Impact
With the 6-8 hour limit, I prioritized:
1. Booking correctness and logic - no race conditions
2. Webhook Resilience - Requests don't fail very often 
3. Coach assignment - Simple algorithm

### What I Would Add With More Time
- **Next 2 hours**: Better logging and error handling
- **Next day**: Fairness-aware distribution algorithm
- **Next week**: Metrics + dashboard

## Part 3: Deployment Strategy

### Local Development
```bash
# How to run locally
docker-compose up -d
```

### Production Deployment (AWS)

#### Infrastructure
- **Compute**: t3.medium EC2 instance
- **Database**: RDS PostgreSQL db.t3.micro
- **Estimated Cost**: $XX/month

#### Deployment Process
1. GitHub Actions for automatic building and pushing images
2. Terraform to provision infrastructure
3. Jenkins for automatic deployments

#### Scaling Strategy
- **Current**: Single instance with a connection pool
- **Next Level**: Horizontal scaling and Redis replacing most DB calls
- **Auto-scaling triggers**: CPU considering Go's performance

### Monitoring Approach

Did not track any metrics

## Part 4: Testing Strategy

### What I Tested
- [x] Race condition prevention
- [x] Webhook idempotency
- [x] No double booking

### Test Coverage

No test coverage.

### How to Run Tests
N/A

## Part 5: Known Limitations & Future Improvements

### Current Limitations
1. Extremely simple coach distribution logic - Certain coaches may heavily be assignment appointments
2. Large number of queries - Availability API takes between 3-5s on local
3. No metrics tracked

### Security Considerations
- Simple API Key validation 
- Using JWT Keys along with storing sensitive data in Vault

### Performance Optimizations Needed
- Utilize Redis cache across multiple queries
- coach_slots table cleanup and regular cron updating slots
