FROM golang:1.23.1-alpine as builder
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/myapp cmd/api/main.go

FROM jrottenberg/ffmpeg:4.4-alpine
#FROM alpine

WORKDIR /app
COPY --from=builder /app/myapp /app/myapp
COPY --from=builder /app/templates /app/templates

EXPOSE 3003
ENTRYPOINT [ "/app/myapp" ]
