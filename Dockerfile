# build stage
FROM golang:1.18.1-alpine3.15 AS build-env
ARG GIT_COMMIT
ARG VERSION
RUN apk --no-cache add build-base git mercurial gcc
ADD . /src
RUN cd /src && go build -ldflags "-X main.GitCommit=${GIT_COMMIT} -X main.Version=${VERSION}" -o baker ./cmd/baker/main.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/baker /app/
ENTRYPOINT ./baker