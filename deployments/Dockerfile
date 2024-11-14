# Stage 1: Build
FROM golang:1.23 AS builder

WORKDIR /app

COPY ./ .

RUN make init
RUN make build CGO_ENABLED=0

FROM alpine:latest

WORKDIR /root/

ARG COPY_EXAMPLES=false

COPY --from=builder /app/dist/uniflow ./uniflow
COPY --from=builder /app/examples ./examples

RUN if [ "$COPY_EXAMPLES" = "false" ]; then rm -rf ./examples; fi

ENTRYPOINT ["./uniflow"]
