FROM golang:1.25.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.23.3  AS deployment
COPY --from=builder /app/build/standalone /standalone