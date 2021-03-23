FROM golang:1.16.2 as builder
WORKDIR /go/planning-poker
COPY go.mod go.sum ./
RUN go mod download

COPY . /go/planning-poker
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/web

FROM alpine:latest
WORKDIR /go/planning-poker
COPY --from=builder /go/planning-poker/ .
CMD ["./app"]

EXPOSE 10080
