KUBECTL=kubectl
INFRA_DIR=k8s/infra
ITEM_DIR=k8s/item
ORDER_DIR=k8s/order

.PHONY: all grafana prometheus tempo item order deploy clean

all: deploy

deploy: prometheus tempo grafana item order

prometheus:
	@echo "Deploying Prometheus..."
	$(KUBECTL) apply -f $(INFRA_DIR)/prometheus/configmap.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/prometheus/deployment.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/prometheus/service.yaml

tempo:
	@echo "Deploying Tempo..."
	$(KUBECTL) apply -f $(INFRA_DIR)/tempo/configmap.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/tempo/pvc.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/tempo/deployment.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/tempo/service.yaml

grafana:
	@echo "Deploying Grafana..."
	$(KUBECTL) apply -f $(INFRA_DIR)/grafana/pvc.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/grafana/deployment.yaml
	$(KUBECTL) apply -f $(INFRA_DIR)/grafana/service.yaml

item:
	@echo "Deploying Item Service and Redis..."
	$(KUBECTL) apply -f $(ITEM_DIR)/redis-item-pvc.yaml
	$(KUBECTL) apply -f $(ITEM_DIR)/redis-item-service.yaml
	$(KUBECTL) apply -f $(ITEM_DIR)/item-service.yaml

order:
	@echo "Deploying Order Service and Redis..."
	$(KUBECTL) apply -f $(ORDER_DIR)/redis-order-pvc.yaml
	$(KUBECTL) apply -f $(ORDER_DIR)/redis-order-service.yaml
	$(KUBECTL) apply -f $(ORDER_DIR)/order-service.yaml

clean:
	@echo "Deleting all resources..."
	$(KUBECTL) delete -f $(ORDER_DIR) || true
	$(KUBECTL) delete -f $(ITEM_DIR) || true
	$(KUBECTL) delete -f $(INFRA_DIR)/grafana || true
	$(KUBECTL) delete -f $(INFRA_DIR)/tempo || true
	$(KUBECTL) delete -f $(INFRA_DIR)/prometheus || true
