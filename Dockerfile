FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main cmd/node/main.go

FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8081
EXPOSE 9091

CMD ["./main"]
