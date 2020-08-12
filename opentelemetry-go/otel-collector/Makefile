
build:
	go build -o main main.go
	docker build -t demo:latest .

start:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down

.PHONY:
	build
