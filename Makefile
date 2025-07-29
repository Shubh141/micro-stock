KUBECTL=kubectl
K8S_DIR=k8s/infra

.PHONY: all grafana prometheus tempo item order deploy clean

all: deploy

deploy: prometheus tempo grafana item order

prometheus:
	@echo "Deploying Prometheus..."
	$(KUBECTL) apply -f $(K8S_DIR)/prometheus/configmap.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/prometheus/deployment.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/prometheus/service.yaml

tempo:
	@echo "Deploying Tempo..."
	$(KUBECTL) apply -f $(K8S_DIR)/tempo/configmap.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/tempo/pvc.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/tempo/deployment.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/tempo/service.yaml

grafana:
	@echo "Deploying Grafana..."
	$(KUBECTL) apply -f $(K8S_DIR)/grafana/pvc.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/grafana/deployment.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/grafana/service.yaml

item:
	@echo "Deploying Item Service and Redis..."
	$(KUBECTL) apply -f $(K8S_DIR)/item/redis-item-pvc.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/item/redis-item-service.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/item/item-service.yaml

order:
	@echo "Deploying Order Service and Redis..."
	$(KUBECTL) apply -f $(K8S_DIR)/order/redis-order-pvc.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/order/redis-order-service.yaml
	$(KUBECTL) apply -f $(K8S_DIR)/order/order-service.yaml

clean:
	@echo "Deleting all resources..."
	$(KUBECTL) delete -f $(K8S_DIR)/order || true
	$(KUBECTL) delete -f $(K8S_DIR)/item || true
	$(KUBECTL) delete -f $(K8S_DIR)/grafana || true
	$(KUBECTL) delete -f $(K8S_DIR)/tempo || true
	$(KUBECTL) delete -f $(K8S_DIR)/prometheus || true
