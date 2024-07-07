import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1, // number of virtual users
  duration: '10s', // duration of the test
};

export default function () {
  const url = 'http://localhost:8090/inmueble';
  const params = {
    page: 1,
    size: 3,
    hasAC: true,
    dateFrom: '2020-12-24',
    dateTo: '2021-12-24'
  };
  const queryString = `page=${params.page}&size=${params.size}&hasAC=${params.hasAC}&dateFrom=${params.dateFrom}&dateTo=${params.dateTo}`;
  const res = http.get(`${url}?${queryString}`);

  // Log the response time
  console.log(`Response time: ${res.timings.duration} ms`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time is less than 500ms': (r) => r.timings.duration < 500,
    'response body is not empty': (r) => r.body.length > 0,
  });

  sleep(1); // wait for 1 second between requests
}
