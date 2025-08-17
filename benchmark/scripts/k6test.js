import http from 'k6/http';
import { check, sleep } from 'k6';

// export const options = {
//     vus: 199,
//     duration: '30s',
// };

export const options = {
    scenarios: {
        performance_test: {
            executor: 'constant-arrival-rate',
            rate: 20000,               // 20k requests per second
            timeUnit: '1s',
            duration: '30s',
            preAllocatedVUs: 1000,
            maxVUs: 2000,
        },
    },
};

export default function () {
    const url = __ENV.TARGET_URL;

    const res = http.get(url, {
        headers: {
            'Connection': 'keep-alive',
        },
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

   // sleep(1);
}