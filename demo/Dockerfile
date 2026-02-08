FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY src .
RUN CGO_ENABLED=0 go build -o app

FROM scratch
LABEL maintainer="dchernenko"
LABEL version="3.0.0"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ADD ./html /html
COPY config.json .
COPY --from=builder /src/app .
ENTRYPOINT ["/app"]
EXPOSE 8088
