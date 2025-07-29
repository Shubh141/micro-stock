# micro-stock

`micro-stock` is a Golang-based microservices application for managing stock and placing orders. Each service maintains its own Redis database. The project includes distributed tracing (OpenTelemetry + Tempo) and monitoring (Prometheus + Grafana). It can be run locally via Docker Compose or deployed to Kubernetes using either raw manifests or a Helm chart.

## Architecture

![System Architecture](images/architecture.png)

This diagram illustrates the core services (item-service and order-service), their connections to Redis, and the observability stack (Prometheus, Tempo, Grafana).


## How to Run
### Option 1: Run Locally with Docker Compose

```bash
docker-compose up --build
```
Available at:

- Item Service: http://localhost:8080

- Order Service: http://localhost:8081

- Prometheus: http://localhost:9090

- Grafana: http://localhost:3000 (admin/admin)

### Option 2: Run on Kubernetes via Makefile
Create a cluster on GKE
```bash
gcloud container clusters create micro-stock-cluster \
  --zone australia-southeast1-b \
  --num-nodes 1 \
  --enable-ip-alias
```
Authenticate kubectl
```bash
gcloud container clusters get-credentials micro-stock-cluster \
  --zone australia-southeast1-b
```
Deploy the system
```bash
make deploy
```
This will deploy the following components in order:

- Prometheus

- Tempo

- Grafana

- Item service and Redis

- Order service and Redis

Clean up all resources:
```bash
make clean
```

Port forward for testing:
```bash
kubectl port-forward svc/item-service 8080:8080
kubectl port-forward svc/order-service 8081:8081
kubectl port-forward svc/grafana 3000:3000
```

### Option 3: Option 3: Deploy with Helm
```bash
helm install micro-stock ./helm-micro-stock
```
To uninstall:
```bash
helm uninstall micro-stock
```

## API Reference

### Item Service (`:8080`)

| Method | Endpoint                    | Description                     |
|--------|-----------------------------|---------------------------------|
| GET    | `/stock/:item`              | Get stock for a specific item   |
| GET    | `/stock`                    | List all items and stock levels |
| POST   | `/stock/:item`              | Set stock for an item           |
| POST   | `/stock/:item/decrement`    | Decrement stock for an item     |
| DELETE | `/stock/:item`              | Delete an item from inventory   |
| GET    | `/healthz`                  | Health check                    |
| GET    | `/metrics`                  | Prometheus metrics              |

### Order Service (`:8081`)

| Method | Endpoint    | Description     |
|--------|-------------|-----------------|
| POST   | `/order`    | Place an order  |
| GET    | `/healthz`  | Health check    |
| GET    | `/metrics`  | Prometheus metrics |


## Observability
### Tracing (OpenTelemetry + Tempo)
Traces are automatically exported to Tempo via OpenTelemetry instrumentation in both services.

Grafana is configured with a Tempo data source:
```bash
http://tempo:3200
```

### Metrics (Prometheus + Grafana)
Both services expose a `/metrics` endpoint, which is scraped by Prometheus and visualised via Grafana.
#### Exported Metrics

| Metric Name                    | Type      | Description                                        | Labels                     |
|-------------------------------|-----------|----------------------------------------------------|----------------------------|
| `http_request_duration_seconds` | Histogram | Duration of HTTP requests                          | `path`, `method`           |
| `http_responses_total`        | Counter   | Count of HTTP responses by status code             | `path`, `method`, `code`   |
| `http_errors_total`           | Counter   | Count of HTTP error responses (status ≥ 400)       | `path`, `method`, `code`   |
| `items_created_total`         | Counter   | Number of items created via `SetStock`             | *(no labels)*              |
| `items_deleted_total`         | Counter   | Number of items deleted via `DeleteStock`          | *(no labels)*              |
| `item_stock_level`            | Gauge     | Current stock level per item                       | `item`                     |
| `orders_placed_total`         | Counter   | Number of orders placed, grouped by item           | `item`                     |

### Grafana Configuration
Grafana is configured with the following data sources:

| Data Source | URL                                    | Purpose              |
|-------------|-----------------------------------------|----------------------|
| Prometheus  | `http://micro-stock-prometheus:9090`   | Metrics scraping     |
| Tempo       | `http://tempo:3200`                    | Trace aggregation    |


## Teardown
To delete the GKE cluster:
```
gcloud container clusters delete micro-stock-cluster \
  --zone australia-southeast1-b
```



