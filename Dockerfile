FROM golang:1.23 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o application main.go

FROM alpine:latest as runner

WORKDIR /car_estimator_auth

COPY --from=builder /build/database/migrations ./database/migrations
COPY --from=builder /build/.env .
COPY --from=builder /build/application .

EXPOSE 4444

CMD ["./application"]
