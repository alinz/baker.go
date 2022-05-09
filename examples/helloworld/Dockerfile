### build stage ###
FROM golang:1.18-alpine AS builder

WORKDIR /helloworld

COPY . .

RUN go build -o server /helloworld/main.go

### run stage ###
FROM alpine:3.12
COPY --from=builder /helloworld/server ./server
CMD ["./server"]