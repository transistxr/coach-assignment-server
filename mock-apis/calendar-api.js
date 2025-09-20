const express = require('express');
const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3001;
const FAILURE_RATE = parseFloat(process.env.FAILURE_RATE || '0.1');

// Simulate random failures
const shouldFail = () => Math.random() < FAILURE_RATE;

// Generate available slots for a coach
function generateAvailableSlots(coachId, days = 7) {
    const slots = [];
    const now = new Date();
    
    for (let day = 0; day < days; day++) {
        const date = new Date(now);
        date.setDate(date.getDate() + day);
        date.setHours(9, 0, 0, 0); // Start at 9 AM
        
        // Skip weekends
        if (date.getDay() === 0 || date.getDay() === 6) continue;
        
        // Generate slots from 9 AM to 5 PM
        for (let hour = 9; hour < 17; hour++) {
            for (let minute = 0; minute < 60; minute += 15) {
                // Randomly mark some slots as unavailable
                if (Math.random() > 0.3) {
                    const slotTime = new Date(date);
                    slotTime.setHours(hour, minute, 0, 0);
                    
                    slots.push({
                        start_time: slotTime.toISOString(),
                        end_time: new Date(slotTime.getTime() + 30 * 60000).toISOString(),
                        available: true
                    });
                }
            }
        }
    }
    
    return slots;
}

// Health check
app.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'mock-calendar-api' });
});

// Get coach availability
app.get('/coaches/:coachId/availability', (req, res) => {
    if (shouldFail()) {
        return res.status(503).json({ 
            error: 'Service temporarily unavailable',
            retry_after: 5
        });
    }
    
    // Simulate processing delay
    setTimeout(() => {
        const { coachId } = req.params;
        const days = parseInt(req.query.days) || 7;
        
        const slots = generateAvailableSlots(coachId, days);
        
        res.json({
            coach_id: coachId,
            timezone: 'America/New_York',
            slots: slots,
            total_available: slots.length,
            generated_at: new Date().toISOString()
        });
    }, Math.random() * 500 + 100); // 100-600ms delay
});

// Block a time slot
app.post('/coaches/:coachId/block-slot', (req, res) => {
    if (shouldFail()) {
        return res.status(503).json({ 
            error: 'Service temporarily unavailable',
            retry_after: 5
        });
    }
    
    const { coachId } = req.params;
    const { start_time, end_time } = req.body;
    
    // Simulate 5% chance of slot already taken
    if (Math.random() < 0.05) {
        return res.status(409).json({
            error: 'Slot already blocked',
            coach_id: coachId,
            start_time,
            end_time
        });
    }
    
    // Simulate processing delay
    setTimeout(() => {
        res.json({
            success: true,
            coach_id: coachId,
            start_time,
            end_time,
            blocked_at: new Date().toISOString(),
            block_id: `block_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        });
    }, Math.random() * 300 + 50); // 50-350ms delay
});

// Release a blocked slot
app.post('/coaches/:coachId/release-slot', (req, res) => {
    const { coachId } = req.params;
    const { block_id } = req.body;
    
    res.json({
        success: true,
        coach_id: coachId,
        block_id,
        released_at: new Date().toISOString()
    });
});

// Get coach calendar settings
app.get('/coaches/:coachId/settings', (req, res) => {
    const { coachId } = req.params;
    
    res.json({
        coach_id: coachId,
        working_hours: {
            start: '09:00',
            end: '17:00',
            timezone: 'America/New_York'
        },
        availability_rules: {
            min_notice_hours: 2,
            max_advance_days: 30,
            buffer_minutes: 5
        },
        blocked_dates: [
            '2024-12-25',
            '2024-12-26',
            '2025-01-01'
        ]
    });
});

// Webhook endpoint for calendar updates
app.post('/webhooks/calendar-update', (req, res) => {
    console.log('Received calendar update webhook:', req.body);
    res.json({ received: true, timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
    console.log(`Mock Calendar API running on port ${PORT}`);
    console.log(`Failure rate: ${FAILURE_RATE * 100}%`);
});