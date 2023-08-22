FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum *.go ./
COPY pkg ./pkg
COPY config ./config

ENV CGO_ENABLED 1
ENV GOOS linux

RUN go build -o manki

CMD ["./manki"]