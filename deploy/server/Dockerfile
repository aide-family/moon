FROM ghcr.io/aide-family/moon:builder AS builder

ARG APP_NAME=""

WORKDIR /moon

COPY ../../ .

RUN make all
RUN make build app=${APP_NAME}

FROM ghcr.io/aide-family/moon:deploy

ARG APP_NAME=""

WORKDIR /moon

COPY --from=builder /moon/bin /moon/bin
COPY --from=builder /moon/cmd/${APP_NAME}/config /moon/config
COPY --from=builder /moon/i18n /moon/i18n

RUN mv /moon/bin/${APP_NAME} /moon/bin/server

VOLUME /moon/config

CMD ["./bin/server", "-c", "/moon/config"]

