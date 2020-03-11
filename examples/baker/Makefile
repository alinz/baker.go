compile:
	GOOS=linux GOARCH=amd64 go build -o baker ../../cmd/baker/main.go

build: compile
	docker rmi baker:latest -f
	docker build -t baker .
	rm baker

run: build
	docker-compose down
	docker-compose up


