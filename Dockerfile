FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /rag-app .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /rag-app /app/rag-app

EXPOSE 8080

CMD ["/app/rag-app"]
