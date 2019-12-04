FROM golang:1.13.4-alpine3.10 AS builder

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -mod vendor -o server .

FROM wpblogger/genpdf-cli:latest

WORKDIR /

COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
