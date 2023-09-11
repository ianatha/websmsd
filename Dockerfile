# syntax=docker/dockerfile:1
FROM golang:1.21
RUN apt update
RUN apt install libusb-1.0-0-dev usbutils usb-modeswitch -y
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN GOOS=linux go build -o /websmsd
EXPOSE 8080
CMD ["/websmsd"]