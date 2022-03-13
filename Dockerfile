FROM golang:1.16.15-bullseye as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build .


RUN apt update && apt install wget -y
RUN wget https://fastdl.mongodb.org/tools/db/mongodb-database-tools-ubuntu2004-x86_64-100.5.2.deb
RUN dpkg -i mongodb-database-tools-ubuntu2004-x86_64-100.5.2.deb

CMD /app/client