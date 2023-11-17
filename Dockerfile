FROM golang:1.21.4 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download 

COPY ./ ./

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -trimpath -o /dist/currency-rates-cronjob
RUN ldd /dist/currency-rates-cronjob | tr -s [:blank:] '\n' | grep ^/ | xargs -I % install -D % /dist/%
RUN ln -s ld-musl-x86_64.so.1 /dist/lib/libc.musl-x86_64.so.1

FROM scratch AS runner

COPY --from=builder /dist /

ENTRYPOINT [ "/currency-rates-cronjob" ]