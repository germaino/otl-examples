
build:
	go build main.go
	#go build demoapp.go
	docker build -t demo:latest .

start:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down

.PHONY:
	build
