FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum *.go ./
COPY data ./data
COPY handler ./handler


ENV CGO_ENABLED 1
ENV GOOS linux

RUN go build -o manki

CMD ["./manki"]
# CMD ["go", "run", "."]