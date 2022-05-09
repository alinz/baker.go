# build stage
FROM golang:1.18-alpine AS builder
ARG GIT_COMMIT
ARG VERSION
RUN apk --no-cache add build-base git mercurial gcc

WORKDIR /baker

COPY . .

RUN go build -ldflags "-X main.GitCommit=${GIT_COMMIT} -X main.Version=${VERSION}" -o server ./cmd/baker/main.go

# final stage
FROM alpine:latest
WORKDIR /baker
COPY --from=builder /baker/server /baker/
ENTRYPOINT ./server