FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /bin/worker ./cmd/worker

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/worker /bin/worker

EXPOSE 8080 9091

CMD ["/bin/api"]
