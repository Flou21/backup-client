FROM flou21/golang:mongo-tools

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build .

CMD /app/client