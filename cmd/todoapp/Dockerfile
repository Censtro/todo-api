FROM golang:1.26.0-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /build/todoapp ./cmd/todoapp \
	&& CGO_ENABLED=0 go build -o /build/task-audit-consumer ./cmd/task-audit-consumer

FROM alpine:3.23
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/todoapp /app/todoapp
COPY --from=builder /build/task-audit-consumer /app/task-audit-consumer
CMD ["/app/todoapp"]

