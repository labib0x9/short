import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter } from 'k6/metrics';

// ---- Config (override via -e FLAG=value on the CLI, see examples below) ----
const POST_URL = __ENV.POST_URL || 'http://127.0.0.1:3000/short';
const POST_RATIO = Number(__ENV.POST_RATIO || 0.2); // 20% POST, 80% GET by default
const RATE = Number(__ENV.RATE || 1000);             // requests per second
const DURATION = __ENV.DURATION || '60s';
const PRE_VUS = Number(__ENV.PRE_VUS || 50);        // pre-allocated VUs (raise if you see "insufficient VUs" warnings)
const MAX_VUS = Number(__ENV.MAX_VUS || 500);       // hard ceiling k6 can scale up to

// ---- Custom metric: count of 500 Internal Server Error responses ----
const serverErrors = new Counter('server_errors_500');

// ---- Load GET URLs from file once, shared read-only across all VUs ----
// get_urls.txt: one full URL per line, mix of:
//   http://127.0.0.1:3000/<code>        (redirect, expect 302)
//   http://127.0.0.1:3000/<code>/stat   (stats endpoint, expect 200)
const getUrls = new SharedArray('get urls', function () {
  return open('./get_urls.txt')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
});

export const options = {
  scenarios: {
    mixed_traffic: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_VUS,
      maxVUs: MAX_VUS,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],   // flag if error rate goes above 1%
    http_req_duration: ['p(95)<500'], // flag if p95 exceeds 500ms
  },
  maxRedirects: 0, // don't auto-follow 3xx redirects (e.g. 302) — inspect the redirect response itself
};

function randomSlug(length = 8) {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  let out = '';
  for (let i = 0; i < length; i++) {
    out += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return out;
}

export default function () {
  const isPost = Math.random() < POST_RATIO;
  let res;

  if (isPost) {
    const fakeUrl = `https://example.com/${randomSlug(6)}/${randomSlug(8)}`;
    const payload = JSON.stringify({ url: fakeUrl });
    const params = { headers: { 'Content-Type': 'application/json' } };

    res = http.post(POST_URL, payload, params);
    check(res, {
      'POST status is 201': (r) => r.status === 201,
    });
  } else {
    const url = getUrls[Math.floor(Math.random() * getUrls.length)];
    res = http.get(url, { redirects: 0 });

    if (url.endsWith('/stat')) {
      check(res, {
        'GET /stat status is 200': (r) => r.status === 200,
      });
    } else {
      check(res, {
        'GET redirect status is 302': (r) => r.status == 302,
      });
    }
  }

  if (res.status === 500) {
    serverErrors.add(1);
  }
}