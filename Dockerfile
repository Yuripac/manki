FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./

COPY *.go ./

RUN CGO_ENABLED=1 GOOS=linux go build -o manki

CMD ["./manki"]