
build:
	go build -o client client/client.go
	go build -o server server/server.go
	docker build -t demo-http-client:latest client
	docker build -t demo-http-server:latest server

start:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down

.PHONY:
	build
