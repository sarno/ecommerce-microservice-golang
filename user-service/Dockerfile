FROM golang:1.25-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /user-service ./cmd/main.go

EXPOSE 8090

CMD [ "/user-service", "db:seed", "start" ]