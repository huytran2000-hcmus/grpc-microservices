.PHONY: db
db:
	docker run -p 3306:3306 \
		 --rm \
		 -e MYSQL_ROOT_PASSWORD=verysecretpass \
		 -v ./init.sql:/docker-entrypoint-initdb.d/my-init.sql \
		 --name grpc-service-dbs \
		 mysql

.PHONY: order
order:
	cd order; DATA_SOURCE_URL="order_service:verysecretpass@tcp(127.0.0.1:3306)/order_service?parseTime=true" \
	APPLICATION_PORT=3000 \
	ENV=development \
	PAYMENT_SERVICE_URL=localhost:3001 \
	OTLP_ENDPOINT=${OTLP_ENDPOINT} \
	go run cmd/main.go

.PHONY: payment
payment:
	cd payment; DATA_SOURCE_URL="payment_service:verysecretpass@tcp(127.0.0.1:3306)/payment_service?parseTime=true" \
	APPLICATION_PORT=3001 \
	ENV=development \
	OTLP_ENDPOINT=${OTLP_ENDPOINT} \
	go run cmd/main.go

.PHONY: docker-compose
docker-compose:
	cd e2e/resources; docker-compose up

.PHONY: mock
mock:
	cd order; mockery

.PHONY: test
test:
	cd order; go test --count=1 --race ./...
	cd e2e; go test --count=1 -v --race ./...

build-images:
	eval $$(minikube docker-env); cd order; docker build -t huypk2000/grpc-order-service:1.0.0 .
	eval $$(minikube docker-env); cd payment; docker build -t huypk2000/grpc-payment-service:1.0.0 .

test-container:
	kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm

helm-env:
	minikube addons enable ingress
	minikube addons enable ingress-dns
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

helm-clean-data:
	kubectl delete pvc --all
	kubectl delete pv --all

helm-install:
	kubectl create secret generic jaeger-secret --from-literal=ES_PASSWORD=admin --from-literal=ES_USERNAME=admin
	helm install grpc-microservices deploy/grpc-microservices/ 
	minikube tunnel

helm-uninstall:
	helm uninstall grpc-microservices
	kubectl delete secret jaeger-secret

helm-upgrade:
	helm upgrade grpc-microservices deploy/grpc-microservices/

grpcurl-local-test-create-order:
	grpcurl --plaintext --import-path ../grpc-microservices-proto/order --proto order.proto -d '{"user_id": 123, "order_items": [{"product_code": "prod","quantity": 4, "unit_price": 12}]}' localhost:3000  Order/Create 

grpcurl-local-test-get-order:
	grpcurl --plaintext --import-path ../grpc-microservices-proto/order --proto order.proto -d '{"order_id": 1}' localhost:3000  Order/Get 

grpcurl-test-create-order:
	grpcurl --insecure --import-path ../grpc-microservices-proto/order --proto order.proto -d '{"user_id": 123, "order_items": [{"product_code": "prod","quantity": 4, "unit_price": 12}]}' ingress.local:443  Order/Create 

grpcurl-test-get-order:
	grpcurl --insecure --import-path ../grpc-microservices-proto/order --proto order.proto -d '{"order_id": 1}' ingress.local:443  Order/Get 

