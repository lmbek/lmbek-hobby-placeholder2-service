FROM golang:1.21-alpine
WORKDIR /app
COPY main.go .
RUN go build -o service main.go
EXPOSE 8081
CMD ["./service"]
