FROM golang:1.25-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /product-service ./cmd/main.go

EXPOSE 8082

CMD [ "/product-service", "start", "worker", "worker:delete-product", "worker:update-stock" ]