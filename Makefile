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
