
build:
	go build main.go
	docker build -t demo-otl-collector-std:latest .

start:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down

.PHONY:
	build
