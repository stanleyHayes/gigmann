// k6 load test for the Daily Brief hot path (GEC-102).
//   k6 run -e BASE_URL=http://localhost:8080 -e EMAIL=ceo@gigmann.health -e PASSWORD=ahenfie-demo infra/load/brief-load.js
import http from 'k6/http'
import { check, sleep } from 'k6'

const BASE = __ENV.BASE_URL || 'http://localhost:8080'
const EMAIL = __ENV.EMAIL || 'ceo@gigmann.health'
const PASSWORD = __ENV.PASSWORD || 'ahenfie-demo'

export const options = {
  stages: [
    { duration: '30s', target: 25 },
    { duration: '1m', target: 25 },
    { duration: '15s', target: 0 },
  ],
  thresholds: {
    // The brief is cached + pre-warmed, so the hot path should be fast.
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
}

export function setup() {
  const res = http.post(`${BASE}/api/v1/auth/login`, JSON.stringify({ email: EMAIL, password: PASSWORD }), {
    headers: { 'Content-Type': 'application/json' },
  })
  check(res, { 'login 200': (r) => r.status === 200 })
  return { token: res.json('token') }
}

export default function (data) {
  const headers = { Authorization: `Bearer ${data.token}` }
  const brief = http.get(`${BASE}/api/v1/brief`, { headers })
  check(brief, { 'brief 200': (r) => r.status === 200 })
  const metrics = http.get(`${BASE}/api/v1/metrics`, { headers })
  check(metrics, { 'metrics 200': (r) => r.status === 200 })
  sleep(1)
}
