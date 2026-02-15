FROM golang:latest AS builder

WORKDIR /usr/local/src

COPY . .

RUN go vet -v ./...

RUN go build -o gshortener cmd/shortener/main.go

FROM alpine:latest

ENV SERVER_ADDRESS=localhost:8080

ENV BASE_URL=http://localhost:8080

WORKDIR /usr/local/bin

COPY --from=builder /usr/local/src/gshortener /usr/local/bin/gshortener

ENTRYPOINT [ "/usr/local/bin/gshortener" ]