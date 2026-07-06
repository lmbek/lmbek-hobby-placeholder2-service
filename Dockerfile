FROM golang:1.26-alpine
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o service main.go
EXPOSE 8081
CMD ["./service"]
