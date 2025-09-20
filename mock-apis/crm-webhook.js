const express = require('express');
const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3002;
const FAILURE_RATE = parseFloat(process.env.FAILURE_RATE || '0.3');

// Store received webhooks for verification
const webhookLog = [];
const MAX_LOG_SIZE = 100;

// Simulate random failures
const shouldFail = () => Math.random() < FAILURE_RATE;

// Health check
app.get('/health', (req, res) => {
    res.json({ 
        status: 'healthy', 
        service: 'mock-crm-webhook',
        failure_rate: FAILURE_RATE,
        webhooks_received: webhookLog.length
    });
});

// Receive appointment created webhook
app.post('/webhooks/appointment-created', (req, res) => {
    // Check for idempotency key
    const idempotencyKey = req.headers['x-idempotency-key'];
    
    if (!idempotencyKey) {
        return res.status(400).json({ 
            error: 'Missing X-Idempotency-Key header' 
        });
    }
    
    // Check if we've seen this key before
    const existingWebhook = webhookLog.find(w => w.idempotencyKey === idempotencyKey);
    if (existingWebhook) {
        console.log(`Duplicate webhook detected: ${idempotencyKey}`);
        return res.json({ 
            success: true,
            message: 'Already processed',
            crm_id: existingWebhook.crmId
        });
    }
    
    // Simulate failures
    if (shouldFail()) {
        console.log(`Simulating failure for webhook: ${idempotencyKey}`);
        return res.status(503).json({ 
            error: 'Service temporarily unavailable',
            retry_after: 5
        });
    }
    
    // Simulate processing delay
    setTimeout(() => {
        const crmId = `crm_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        
        // Log the webhook
        const webhookEntry = {
            idempotencyKey,
            crmId,
            type: 'appointment-created',
            payload: req.body,
            received_at: new Date().toISOString()
        };
        
        webhookLog.push(webhookEntry);
        if (webhookLog.length > MAX_LOG_SIZE) {
            webhookLog.shift(); // Remove oldest entry
        }
        
        console.log('Appointment created webhook processed:', {
            appointment_id: req.body.appointment_id,
            crm_id: crmId
        });
        
        res.json({
            success: true,
            crm_id: crmId,
            processed_at: new Date().toISOString()
        });
    }, Math.random() * 1000 + 200); // 200-1200ms delay
});

// Receive appointment updated webhook
app.post('/webhooks/appointment-updated', (req, res) => {
    const idempotencyKey = req.headers['x-idempotency-key'];
    
    if (!idempotencyKey) {
        return res.status(400).json({ 
            error: 'Missing X-Idempotency-Key header' 
        });
    }
    
    // Check for duplicate
    const existingWebhook = webhookLog.find(w => w.idempotencyKey === idempotencyKey);
    if (existingWebhook) {
        return res.json({ 
            success: true,
            message: 'Already processed'
        });
    }
    
    // Simulate failures
    if (shouldFail()) {
        return res.status(503).json({ 
            error: 'Service temporarily unavailable',
            retry_after: 5
        });
    }
    
    // Process webhook
    setTimeout(() => {
        const webhookEntry = {
            idempotencyKey,
            type: 'appointment-updated',
            payload: req.body,
            received_at: new Date().toISOString()
        };
        
        webhookLog.push(webhookEntry);
        if (webhookLog.length > MAX_LOG_SIZE) {
            webhookLog.shift();
        }
        
        console.log('Appointment updated webhook processed:', {
            appointment_id: req.body.appointment_id,
            status: req.body.status
        });
        
        res.json({
            success: true,
            processed_at: new Date().toISOString()
        });
    }, Math.random() * 800 + 100);
});

// Receive appointment cancelled webhook
app.post('/webhooks/appointment-cancelled', (req, res) => {
    const idempotencyKey = req.headers['x-idempotency-key'];
    
    if (!idempotencyKey) {
        return res.status(400).json({ 
            error: 'Missing X-Idempotency-Key header' 
        });
    }
    
    // Simulate occasional timeout (no response)
    if (Math.random() < 0.05) {
        console.log('Simulating timeout for webhook');
        // Don't send response - let it timeout
        return;
    }
    
    if (shouldFail()) {
        return res.status(503).json({ 
            error: 'Service temporarily unavailable',
            retry_after: 5
        });
    }
    
    res.json({
        success: true,
        processed_at: new Date().toISOString()
    });
});

// Get webhook log (for debugging/testing)
app.get('/webhooks/log', (req, res) => {
    res.json({
        total_received: webhookLog.length,
        webhooks: webhookLog.slice(-20) // Last 20 webhooks
    });
});

// Clear webhook log (for testing)
app.delete('/webhooks/log', (req, res) => {
    webhookLog.length = 0;
    res.json({ message: 'Webhook log cleared' });
});

// Simulate CRM contact lookup
app.get('/contacts/:contactId', (req, res) => {
    const { contactId } = req.params;
    
    if (shouldFail()) {
        return res.status(503).json({ 
            error: 'Service temporarily unavailable'
        });
    }
    
    // Simulate contact not found
    if (Math.random() < 0.1) {
        return res.status(404).json({ 
            error: 'Contact not found' 
        });
    }
    
    res.json({
        id: contactId,
        email: `contact${contactId}@example.com`,
        name: `Contact ${contactId}`,
        phone: '+1234567890',
        tags: ['lead', 'scheduled'],
        score: Math.random() * 100,
        created_at: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
    });
});

app.listen(PORT, () => {
    console.log(`Mock CRM Webhook running on port ${PORT}`);
    console.log(`Failure rate: ${FAILURE_RATE * 100}%`);
});