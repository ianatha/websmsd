# syntax=docker/dockerfile:1
FROM golang:1.19
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
ADD modem ./modem
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /websmsd
EXPOSE 3000
CMD ["/websmsd"]