winget install k6 --source winget

## Optional

winget install --id Cloudflare.cloudflared
cloudflared tunnel --url http://localhost:8090

## To run

k6 run file.js


     ✓ status is 200
     ✗ response time is less than 1000ms
      ↳  56% — ✓ 114 / ✗ 88
     ✓ response body is not empty

     █ setup

       ✓ response time is less than 500ms

     checks.........................: 85.52% ✓ 520       ✗ 88
     data_received..................: 52 kB  4.9 kB/s
     data_sent......................: 49 kB  4.6 kB/s
     http_req_blocked...............: avg=21.93ms  min=0s       med=0s      max=108.74ms p(90)=59ms     p(95)=66.84ms
     http_req_connecting............: avg=21.57ms  min=0s       med=0s      max=70.99ms  p(90)=58.99ms  p(95)=65.85ms
     http_req_duration..............: avg=2.81s    min=25ms     med=457.1ms max=10.19s   p(90)=8.29s    p(95)=9.25s
       { expected_response:true }...: avg=2.83s    min=129.32ms med=466.7ms max=10.19s   p(90)=8.3s     p(95)=9.26s
     http_req_failed................: 0.98%  ✓ 2         ✗ 202
     http_req_receiving.............: avg=506.33µs min=0s       med=0s      max=44.22ms  p(90)=999.47µs p(95)=1.45ms
     http_req_sending...............: avg=644.51µs min=0s       med=0s      max=4.01ms   p(90)=2.99ms   p(95)=3ms
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=2.8s     min=24.44ms  med=454.6ms max=10.19s   p(90)=8.29s    p(95)=9.25s
     http_reqs......................: 204    19.074386/s
     iteration_duration.............: avg=5.86s    min=110.71ms med=6.16s   max=10.49s   p(90)=10.11s   p(95)=10.32s
     iterations.....................: 101    9.443691/s
     vus............................: 14     min=14      max=100
     vus_max........................: 100    min=100     max=100


running (10.7s), 000/100 VUs, 101 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1s


