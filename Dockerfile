FROM alpine:latest

WORKDIR /app
COPY ./baker /app/baker

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/app/baker"]