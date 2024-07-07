Following
https://hadlakmal.medium.com/golang-metrics-with-prometheus-and-grafana-db15d1b3b1f8

First time:

docker network create metrics

To start:
docker-compose up

Metrics are exposed
http://127.0.0.1:8181/metrics

Find your ipv4 local ip, and put it in the prometheus.yml in the static_configs:

Prometheus
http://localhost:9090/

Finally Grafana Dashboard, set database as a Prometheus database with the aforementioned url
http://localhost:3000/