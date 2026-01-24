FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o scaletestapp main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/scaletestapp .
EXPOSE 8080
CMD ["./scaletestapp"]