-- Simplified Schema for Scheduling System Challenge
-- This is a subset of the production schema focused on core scheduling functionality

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Coaches table - stores coach information and performance metrics
CREATE TABLE coaches (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    name VARCHAR NOT NULL,
    email VARCHAR UNIQUE NOT NULL,
    score FLOAT DEFAULT 0.0,
    max_daily_appointments INTEGER DEFAULT 10,
    working_hours_start TIME DEFAULT '09:00:00',
    working_hours_end TIME DEFAULT '17:00:00',
    timezone VARCHAR DEFAULT 'America/New_York',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Calendars table - different calendar types/configurations
CREATE TABLE calendars (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    name VARCHAR NOT NULL,
    calendar_type VARCHAR DEFAULT 'standard',
    slot_duration INTEGER DEFAULT 30, -- in minutes
    slot_interval INTEGER DEFAULT 15, -- interval between slot start times - one slot starts at 10AM, the next at 10:15, but 10AM being booked takes out 10:15 for the same coach as well since duration is 30 minutes
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Coach-Calendar relationship mapping - who is on what calendar? This decides teams for each calendar
CREATE TABLE coach_calendars (
    coach_id VARCHAR REFERENCES coaches(id) ON DELETE CASCADE,
    calendar_id VARCHAR REFERENCES calendars(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (coach_id, calendar_id)
);

-- Coach availability slots
CREATE TABLE coach_slots (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    coach_id VARCHAR REFERENCES coaches(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    available BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    UNIQUE(coach_id, start_time)
);

-- Appointments table
CREATE TABLE coach_appointments (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    coach_id VARCHAR REFERENCES coaches(id),
    calendar_id VARCHAR REFERENCES calendars(id),
    contact_id VARCHAR NOT NULL, --the customer's id, don't worry about their db storage right now
    title VARCHAR,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no_show', 'rescheduled')),
    timezone VARCHAR DEFAULT 'UTC',
    
    -- Tracking fields
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,

    -- Integration fields
    external_calendar_id VARCHAR,
    crm_contact_id VARCHAR,
    webhook_status VARCHAR DEFAULT 'pending',
    webhook_attempts INTEGER DEFAULT 0,
    webhook_last_attempt TIMESTAMPTZ,
    
    -- Metadata
    source VARCHAR, -- 'api', 'webhook', 'manual'
    notes TEXT,
    metadata JSONB,
    
    -- Prevent double booking
    CONSTRAINT no_double_booking UNIQUE (coach_id, start_time)
);

-- Webhook events log
CREATE TABLE webhook_events (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    event_type VARCHAR NOT NULL,
    event_source VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    last_attempt TIMESTAMPTZ,
    error_message TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- API keys for authentication
CREATE TABLE api_keys (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    key_hash VARCHAR UNIQUE NOT NULL,
    name VARCHAR NOT NULL,
    active BOOLEAN DEFAULT true,
    rate_limit INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP
);

-- Appointment distribution log (for tracking distribution decisions) - appointment distribution is the process of assigning a certain customer's appointment to a certain coach
CREATE TABLE distribution_log (
    id VARCHAR PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    appointment_id VARCHAR REFERENCES coach_appointments(id),
    coaches_considered JSONB, -- Array of coach IDs with scores
    selected_coach_id VARCHAR REFERENCES coaches(id),
    selection_reason VARCHAR,
    distribution_score FLOAT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_coach_appointments_start_time ON coach_appointments(start_time);
CREATE INDEX idx_coach_appointments_coach_id ON coach_appointments(coach_id);
CREATE INDEX idx_coach_appointments_status ON coach_appointments(status);
CREATE INDEX idx_coach_appointments_calendar_id ON coach_appointments(calendar_id);
CREATE INDEX idx_coach_slots_coach_id ON coach_slots(coach_id);
CREATE INDEX idx_coach_slots_start_time ON coach_slots(start_time);
CREATE INDEX idx_coach_slots_available ON coach_slots(available);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at);

-- Insert sample data
INSERT INTO coaches (id, name, email, score) VALUES
    ('coach-1', 'Alice Johnson', 'alice@example.com', 0.85),
    ('coach-2', 'Bob Smith', 'bob@example.com', 0.72),
    ('coach-3', 'Carol Davis', 'carol@example.com', 0.90),
    ('coach-4', 'David Wilson', 'david@example.com', 0.65),
    ('coach-5', 'Eve Martinez', 'eve@example.com', 0.58),
    ('coach-6', 'Frank Thompson', 'frank@example.com', 0.78);

INSERT INTO calendars (id, name, calendar_type, slot_duration, slot_interval) VALUES
    ('cal-1', 'Sales Consultation', 'standard', 30, 15),
    ('cal-2', 'Technical Interview', 'standard', 45, 15),
    ('cal-3', 'Quick Check-in', 'express', 15, 15);

-- Map coaches to calendars
INSERT INTO coach_calendars (coach_id, calendar_id) VALUES
    ('coach-1', 'cal-1'),
    ('coach-1', 'cal-3'),
    ('coach-2', 'cal-1'),
    ('coach-2', 'cal-2'),
    ('coach-3', 'cal-1'),
    ('coach-3', 'cal-3'),
    ('coach-4', 'cal-2'),
    ('coach-5', 'cal-1'),
    ('coach-6', 'cal-1'),
    ('coach-6', 'cal-2');

-- Insert sample API keys (in practice, these would be hashed)
INSERT INTO api_keys (key_hash, name, rate_limit) VALUES
    ('test-key-123', 'Test API Key', 100),
    ('dev-key-456', 'Development Key', 500),
    ('prod-key-789', 'Production Key', 1000);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_coaches_updated_at BEFORE UPDATE ON coaches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_calendars_updated_at BEFORE UPDATE ON calendars
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_coach_appointments_updated_at BEFORE UPDATE ON coach_appointments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_coach_slots_updated_at BEFORE UPDATE ON coach_slots
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();