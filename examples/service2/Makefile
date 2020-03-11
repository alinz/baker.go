compile:
	GOOS=linux GOARCH=amd64 go build -o ./service2 main.go

build: compile
	docker rmi baker/service2:latest -f
	docker build -t baker/service2 .
	rm service2

run: build
	docker-compose down
	docker-compose up
