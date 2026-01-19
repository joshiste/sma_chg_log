FROM alpine:3.23

ARG TARGETPLATFORM

RUN apk add --no-cache ca-certificates

COPY ${TARGETPLATFORM}/sma_chg_log /usr/local/bin/sma_chg_log

ENTRYPOINT ["sma_chg_log"]
