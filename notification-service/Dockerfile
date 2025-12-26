FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /notification-service ./cmd/main.go

EXPOSE 8081

CMD [ "/notification-service", "db:migrate", "notification:start" ]