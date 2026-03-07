FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-service ./cmd/api

FROM scratch

COPY --from=build /app/api-service /api-service

EXPOSE 8080

ENTRYPOINT ["/api-service"]

