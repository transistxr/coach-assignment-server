const express = require('express');
const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3003;
const VALID_API_KEYS = (process.env.VALID_API_KEYS || 'test-key-123,dev-key-456,prod-key-789').split(',');

// Rate limiting store (in production, use Redis)
const rateLimitStore = new Map();

// Health check
app.get('/health', (req, res) => {
    res.json({ 
        status: 'healthy', 
        service: 'mock-auth-service',
        valid_keys_count: VALID_API_KEYS.length
    });
});

// Validate API key
app.post('/validate', (req, res) => {
    const { api_key } = req.body;
    const apiKeyHeader = req.headers['x-api-key'];
    
    const keyToValidate = api_key || apiKeyHeader;
    
    if (!keyToValidate) {
        return res.status(400).json({ 
            valid: false,
            error: 'Missing API key'
        });
    }
    
    // Check if key is valid
    const isValid = VALID_API_KEYS.includes(keyToValidate);
    
    if (!isValid) {
        return res.status(401).json({ 
            valid: false,
            error: 'Invalid API key'
        });
    }
    
    // Check rate limit
    const now = Date.now();
    const windowMs = 60000; // 1 minute
    const maxRequests = keyToValidate.includes('test') ? 100 : 
                       keyToValidate.includes('dev') ? 500 : 1000;
    
    // Get or create rate limit entry
    if (!rateLimitStore.has(keyToValidate)) {
        rateLimitStore.set(keyToValidate, { count: 0, resetTime: now + windowMs });
    }
    
    const rateLimit = rateLimitStore.get(keyToValidate);
    
    // Reset if window has passed
    if (now > rateLimit.resetTime) {
        rateLimit.count = 0;
        rateLimit.resetTime = now + windowMs;
    }
    
    rateLimit.count++;
    
    // Check if rate limit exceeded
    if (rateLimit.count > maxRequests) {
        return res.status(429).json({ 
            valid: false,
            error: 'Rate limit exceeded',
            retry_after: Math.ceil((rateLimit.resetTime - now) / 1000),
            limit: maxRequests,
            remaining: 0,
            reset: new Date(rateLimit.resetTime).toISOString()
        });
    }
    
    // Return validation result
    res.json({
        valid: true,
        key_type: keyToValidate.includes('test') ? 'test' : 
                  keyToValidate.includes('dev') ? 'development' : 'production',
        rate_limit: {
            limit: maxRequests,
            remaining: maxRequests - rateLimit.count,
            reset: new Date(rateLimit.resetTime).toISOString()
        },
        permissions: {
            read: true,
            write: true,
            delete: keyToValidate.includes('prod')
        }
    });
});

// Get API key info (without validating)
app.get('/keys/:keyId/info', (req, res) => {
    const { keyId } = req.params;
    
    if (!VALID_API_KEYS.includes(keyId)) {
        return res.status(404).json({ 
            error: 'Key not found' 
        });
    }
    
    res.json({
        key_id: keyId,
        type: keyId.includes('test') ? 'test' : 
              keyId.includes('dev') ? 'development' : 'production',
        created_at: '2024-01-01T00:00:00Z',
        last_used: new Date().toISOString(),
        rate_limit: keyId.includes('test') ? 100 : 
                   keyId.includes('dev') ? 500 : 1000,
        active: true
    });
});

// Rotate API key (mock implementation)
app.post('/keys/:keyId/rotate', (req, res) => {
    const { keyId } = req.params;
    
    if (!VALID_API_KEYS.includes(keyId)) {
        return res.status(404).json({ 
            error: 'Key not found' 
        });
    }
    
    const newKey = `${keyId}-rotated-${Date.now()}`;
    
    res.json({
        old_key: keyId,
        new_key: newKey,
        rotated_at: new Date().toISOString(),
        expires_in: 3600 // Old key expires in 1 hour
    });
});

// Clean up rate limit store periodically
setInterval(() => {
    const now = Date.now();
    for (const [key, value] of rateLimitStore.entries()) {
        if (now > value.resetTime + 300000) { // 5 minutes after reset
            rateLimitStore.delete(key);
        }
    }
}, 60000); // Check every minute

app.listen(PORT, () => {
    console.log(`Mock Auth Service running on port ${PORT}`);
    console.log(`Valid API keys: ${VALID_API_KEYS.length}`);
});