
import http from 'k6/http';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  vus: 200,  // Start with fewer VUs
  iterations: 20000,
};

export default function() {
  // Use UUID for guaranteed uniqueness
  const uniqueId = uuidv4();
  const shortId = uniqueId.substring(0, 8);

  const payload = JSON.stringify({
    name: `User-${shortId}`,
    email: `test-${uniqueId}@example.com`,  // Full UUID in email
    provider: 'google',
    providerId: `test-k6-${uniqueId}`,   // Full UUID in provider_id
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': uniqueId
    },
    tags: { name: 'create-user-test' }
  };

  const res = http.post('http://localhost:7000/api/v1.0/auth/test', payload, params);


  // Check if response says "already exists" for our unique data
  if (res.body.includes('already exists') || res.body.includes('duplicate')) {
    console.error(`🚨 BUG DETECTED! Duplicate error for UNIQUE data: ${uniqueId}`);
  }
}
