.PHONY: db
db:
	docker run -p 3306:3306 \
		 -e MYSQL_ROOT_PASSWORD=verysecretpass \
		 -v ./init.sql:/docker-entrypoint-initdb.d/my-init.sql \
		 --name grpc-service-dbs \
		 mysql

.PHONY: order
order:
	cd order; DATA_SOURCE_URL="order_service:verysecretpass@tcp(127.0.0.1:3306)/order_service" \
	APPLICATION_PORT=3000 \
	ENV=development \
	PAYMENT_SERVICE_URL=localhost:3001 \
	go run cmd/main.go

.PHONY: payment
payment:
	cd payment; DATA_SOURCE_URL="payment_service:verysecretpass@tcp(127.0.0.1:3306)/payment_service" \
	APPLICATION_PORT=3001 \
	ENV=development \
	go run cmd/main.go

.PHONY: mock
mock:
	cd order; mockery

.PHONY: test
test:
	cd order; go test --count=1 --race ./...
	cd e2e; go test --count=1 -v --race ./...

build-images:
	eval $(minikube docker-env)
	cd order; docker build -t huypk2000/grpc-order-service:1.0.0 .
	cd payment; docker build -t huypk2000/grpc-payment-service:1.0.0 .

test-container:
	kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm

helm-install:
	helm install grpc-microservices deploy/grpc-microservices/ 
	minikube tunnel

helm-uninstall:
	helm uninstall grpc-microservices

helm-upgrade:
	helm upgrade grpc-microservices deploy/grpc-microservices/

grpcurl-test-request:
	grpcurl --insecure --import-path ../grpc-microservices-proto/order --proto order.proto -d '{"user_id": 123, "order_items": [{"product_code": "prod","quantity": 4, "unit_price": 12}]}' ingress.local:443  Order/Create 

