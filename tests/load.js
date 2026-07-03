import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter, Rate } from 'k6/metrics';

const POST_URL = __ENV.POST_URL || 'http://127.0.0.1:3000/short';
const POST_RATIO = Number(__ENV.POST_RATIO || 0.2);
const RATE = Number(__ENV.RATE || 1000);
const DURATION = __ENV.DURATION || '60s';
const PRE_VUS = Number(__ENV.PRE_VUS || 150);
const MAX_VUS = Number(__ENV.MAX_VUS || 500);

// ---- Per-status-code counters ----
const status200 = new Counter('status_200');
const status201 = new Counter('status_201');
const status302 = new Counter('status_302');
const status400 = new Counter('status_400');
const status404 = new Counter('status_404');
const status410 = new Counter('status_410');
const status429 = new Counter('status_429');
const status500 = new Counter('status_500');
const statusOther = new Counter('status_other'); // catch-all for anything unexpected

// Custom failure rate: only counts responses that are NOT expected business
// outcomes (404/410/429/2xx-per-endpoint). This replaces k6's built-in
// http_req_failed, which flags every non-2xx/3xx response as a "failure"
// even when 404/429 are the correct, intended response for this workload.
const unexpectedFailRate = new Rate('unexpected_fail_rate');

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
    // Only fail the run on genuinely unexpected responses (500s, unknown codes)
    unexpected_fail_rate: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
  maxRedirects: 0,
};

function randomSlug(length = 8) {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  let out = '';
  for (let i = 0; i < length; i++) {
    out += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return out;
}

// Track every response status code seen, regardless of endpoint
function recordStatus(status) {
  switch (status) {
    case 200: status200.add(1); break;
    case 201: status201.add(1); break;
    case 302: status302.add(1); break;
    case 400: status400.add(1); break;
    case 404: status404.add(1); break;
    case 410: status410.add(1); break;
    case 429: status429.add(1); break;
    case 500: status500.add(1); break;
    default: statusOther.add(1); break;
  }
}

const REDIRECT_OK = new Set([302, 404, 410, 429]);
const STAT_OK = new Set([200, 404, 410, 429]);
const CREATE_OK = new Set([201, 429]);

// Tell k6 itself which status codes are "expected" so its built-in
// http_req_failed metric stops flagging every 404/429 as a failure.
// This is the union of all three endpoints' OK sets - k6's callback
// applies globally, not per-request, so per-endpoint correctness is
// still enforced separately via check() + unexpectedFailRate below.
http.setResponseCallback(
  http.expectedStatuses(200, 201, 302, 404, 410, 429)
);

export default function () {
  const isPost = Math.random() < POST_RATIO;
  let res;
  let ok;

  if (isPost) {
    const fakeUrl = `https://example.com/${randomSlug(6)}/${randomSlug(8)}`;
    const payload = JSON.stringify({ url: fakeUrl });
    const params = { headers: { 'Content-Type': 'application/json' } };
    res = http.post(POST_URL, payload, params);
    ok = CREATE_OK.has(res.status);
    check(res, { 'POST /short': () => ok });
  } else {
    const url = getUrls[Math.floor(Math.random() * getUrls.length)];
    res = http.get(url, { redirects: 0 });
    if (url.endsWith('/stat')) {
      ok = STAT_OK.has(res.status);
      check(res, { 'GET /{code}/stat': () => ok });
    } else {
      ok = REDIRECT_OK.has(res.status);
      check(res, { 'GET /{code}': () => ok });
    }
  }

  recordStatus(res.status);
  // A response is only an "unexpected failure" if it's outside the
  // endpoint's own OK set (e.g. 500, or a code not in REDIRECT_OK/STAT_OK/CREATE_OK).
  unexpectedFailRate.add(!ok);
}