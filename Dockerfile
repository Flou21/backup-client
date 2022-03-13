FROM golang:1.16.6-alpine3.14 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build .


FROM alpine:3.14

COPY --from=builder /app/client /usr/local/bin/client

CMD client