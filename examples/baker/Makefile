compile:
	GOOS=linux GOARCH=amd64 go build -o baker ../../cmd/baker/main.go

build: compile
	docker rmi balinz/baker.go:latest -f
	docker build -t alinz/baker.go:latest .
	rm baker

run: build
	docker-compose down
	docker-compose up


