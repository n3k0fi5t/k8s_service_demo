FROM golang:1.13-alpine

ADD . /src
WORKDIR /src

RUN go build -o /src/app main.go

ENTRYPOINT /src/deploy/entrypoint.sh