FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum *.go ./
COPY data ./data
COPY db ./db
COPY handler ./handler
COPY www ./www


ENV CGO_ENABLED 1
ENV GOOS linux

RUN go build -o manki

CMD ["./manki"]
# CMD ["go", "run", "."]