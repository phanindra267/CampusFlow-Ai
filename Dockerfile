FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=UTC
WORKDIR /app
COPY --from=builder /api .
EXPOSE 8080
USER nobody
CMD ["./api"]