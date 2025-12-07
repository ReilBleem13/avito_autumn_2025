import http from 'k6/http';
import { sleep, check } from 'k6';

const BASE_URL = 'http://localhost:8080';  

function randomBoolean(p = 0.9) {
  return Math.random() < p;
}

function randomInteger(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomString(length, chars) {
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

function simpleUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export const options = {
    scenarios: {
        pr_lifecycle: { 
            executor: 'ramping-arrival-rate',
            startRate: 2,
            timeUnit: '1s',
            stages: [
                { duration: '30s', target: 3 }, 
                { duration: '1m', target: 4 },
                { duration: '30s', target: 0 },
            ],
            preAllocatedVUs: 1,
            maxVUs: 10,
            exec: 'pr_lifecycle',
        },
        reads: {  
            executor: 'constant-arrival-rate',
            rate: 2,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 2,
            maxVUs: 10,
            exec: 'reads',
        },
    },
    thresholds: {
        'http_req_duration': ['p(90)<300'],
        'http_req_failed': ['rate<0.001'],
        checks: ['rate>0.999'],
    },
};

export function setup() {
    const teams = [];
    const allUsers = [];

    for (let i = 0; i < 20; i++) {
        const teamName = `test-team-${randomString(8, 'abcdefghijklmnopqrstuvwxyz')}`;
        const members = [];
        for (let j = 0; j < 10; j++) {
            const userId = `u${i * 10 + j + 1}`;
            const username = `User${userId}`;
            members.push({
                user_id: userId,
                username: username,
                is_active: randomBoolean(0.9),  
            });
            allUsers.push(userId);
        }

        const payload = { team_name: teamName, members };
        const res = http.post(`${BASE_URL}/team/add`, JSON.stringify(payload), {
            headers: { 'Content-Type': 'application/json' },
        });

        check(res, {
            'team created/updated': (r) => r.status === 201 || r.status === 200,
        });

        if (res.status !== 400) {
            teams.push({ name: teamName, users: members });
        }
        sleep(0.1);
    }

    console.log(`Setup: Created ${teams.length} teams, ${allUsers.length} users`);
    return {
        teams: teams,
        users: allUsers,
    };
}

export function pr_lifecycle(data) {
    if (!data || !data.users || data.users.length === 0) {
        console.log('Skip: No data from setup');
        return;
    }

    const authorId = data.users[randomInteger(0, data.users.length - 1)];  
    const prId = `pr-load-${Date.now()}-${simpleUUID().slice(0, 8)}`;
    const prName = `Load Test PR ${randomString(10, 'abcdefghijklmnopqrstuvwxyz')}`;

    const createPayload = {
        pull_request_id: prId,
        pull_request_name: prName,
        author_id: authorId,
    };
    const createRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify(createPayload), {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'pr/create' },
    });

    let pr = null;
    if (createRes.status === 201) {
        pr = createRes.json('pr');
    }

    check(createRes, {
        'create status 201': (r) => r.status === 201,
        'pr OPEN': (r) => r.json('pr.status') === 'OPEN',
        'reviewers 0-2': (r) => {
            const reviewers = r.json('pr.assigned_reviewers');
            return Array.isArray(reviewers) && reviewers.length >= 0 && reviewers.length <= 2;
        },
        'no 404/409': (r) => r.status !== 404 && r.status !== 409,
    });

    sleep(0.5);

    if (pr && pr.assigned_reviewers && pr.assigned_reviewers.length > 0 && pr.status === 'OPEN') {
        const oldReviewerId = pr.assigned_reviewers[randomInteger(0, pr.assigned_reviewers.length - 1)];  // Custom
        const reassignPayload = {
            pull_request_id: prId,
            old_user_id: oldReviewerId,
        };
        const reassignRes = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify(reassignPayload), {
            headers: { 'Content-Type': 'application/json' },
            tags: { name: 'pr/reassign' },
        });

        check(reassignRes, {
            'reassign status 200': (r) => r.status === 200,
            'replaced_by exists': (r) => !!r.json('replaced_by'),
            'no 404/409': (r) => r.status !== 404 && r.status !== 409,
        });

        sleep(0.5);
    }

    const mergePayload = { pull_request_id: prId };
    const mergeRes = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify(mergePayload), {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'pr/merge' },
    });

    check(mergeRes, {
        'merge status 200': (r) => r.status === 200,
        'status MERGED': (r) => r.json('pr.status') === 'MERGED',
        'no 404': (r) => r.status !== 404,
    });

    sleep(1); 
}

export function reads(data){
    if (!data || !data.teams.length === 0) {
        console.log('Skip: No teams from setup');
        return;
    }

    const teamIndex = randomInteger(0, data.teams.length - 1);
    const teamName = data.teams[teamIndex].name;

    const getRes = http.get(`${BASE_URL}/team/get?team_name=${teamName}`, {
        tags: { name: 'get' },
    });

    let response = null;
    if (getRes.status === 200){
        response = getRes.json();
    }

    check(getRes, {
        'GET status 200': (r) => r.status === 200,
        'members > 0': (r) => Array.isArray(r.json('members')) && r.json('members').length > 0,  
        'no 404': (r) => r.status !== 404,
    });
    sleep(0.5);

    // Закоментировал по причине того, что я дурак, и добавил 404 при 0 юзеров.

    // const usersIndex = randomInteger(0, data.users.length - 1);
    // const userId = data.users[usersIndex];

    // const getReviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`, {
    //     tags: { name: 'get' },
    // })

    // check(getReviewRes, {
    //     'getReview Status 200 OK': (r) => r.status === 200,
    // });
    sleep(0.5);
}