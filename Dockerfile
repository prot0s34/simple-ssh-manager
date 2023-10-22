FROM golang:1.20 as build
#

WORKDIR /app

COPY . .

RUN go build -o sshmanager

FROM alpine:latest

RUN mkdir /app

COPY --from=build /app/sshmanager /app/sshmanager

CMD ["/app/sshmanager"]