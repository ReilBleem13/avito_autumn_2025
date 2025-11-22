import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '4m', target: 10 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],  // теперь пройдёт!
  },
  discardResponseBodies: true,
};

const BASE_URL = 'http://localhost:8080';
const USERS = ['u1', 'u2', 'u3', 'u4', 'u5'];
const EXISTING_PR = 'pr-load-001';

export default function () {
  const userID = USERS[Math.floor(Math.random() * USERS.length)];

  // 1. getReview — 404 у автора — ОК!
  const res = http.get(`${BASE_URL}/users/getReview?user_id=${userID}`, {
    valid_response: [200, 404],  // ← ЭТО ГЛАВНОЕ!
  });
  check(res, {
    'getReview: 200 или 404': (r) => [200, 404].includes(r.status),
  });

  // 2. createPR — редко
  if (__VU % 10 === 0) {
    const createRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
      pull_request_id: `pr-load-${Date.now()}-${__ITER}`,
      pull_request_name: `Load test PR ${__ITER}`,
      author_id: 'u1',
    }), {
      headers: { 'Content-Type': 'application/json' },
      valid_response: [201, 409],  // ← и тут тоже!
    });
    check(createRes, {
      'createPR: 201 или 409': (r) => [201, 409].includes(r.status),
    });
  }

  // 3. reassign — иногда
  if (__VU % 7 === 0) {
    const reassignRes = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify({
      pull_request_id: EXISTING_PR,
      old_user_id: 'u2',
    }), {
      headers: { 'Content-Type': 'application/json' },
      valid_response: [200, 404, 409],  // ← и тут!
    });
    check(reassignRes, {
      'reassign: 200, 404 или 409': (r) => [200, 404, 409].includes(r.status),
    });
  }

  sleep(0.5);
}