compile:
	GOOS=linux GOARCH=amd64 go build -o ./service main.go

build: compile
	docker rmi baker/service:latest -f
	docker build -t baker/service .
	rm service

run: build
	docker-compose down
	docker-compose up
