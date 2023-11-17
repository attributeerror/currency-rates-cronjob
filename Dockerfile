FROM golang:1.21.4-alpine3.18 AS build-env 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 GOOS=linux go build -o ./currency-rates-cronjob

FROM scratch

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /app/currency-rates-cronjob /currency-rates-cronjob

ENTRYPOINT ["/currency-rates-cronjob"]