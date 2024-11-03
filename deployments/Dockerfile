FROM golang:1.23 as builder

WORKDIR /app

COPY ./ .

RUN make init
RUN make build CGO_ENABLED=0

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/dist/uniflow ./uniflow

ENTRYPOINT ["./uniflow"]
