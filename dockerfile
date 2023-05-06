FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o main.go .

EXPOSE 8080

CMD ["/app/main"]
