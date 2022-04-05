FROM golang:1.17-alpine AS builder

COPY . /code/
WORKDIR /code/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /p2p-chat ./


FROM debian:stretch

COPY --from=builder /p2p-chat /usr/local/bin/p2p-chat
COPY .docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV P2P_CHAT_USERNAME="none"
ENV P2P_CHAT_HOST="127.0.0.1"
ENV P2P_CHAT_PORT="8888"

#ENTRYPOINT [ "p2p-chat", "cli", "-u", "${P2P_CHAT_USERNAME}", "-H", "${P2P_CHAT_HOST}", "-P", "$P2P_CHAT_PORT" ]
#ENTRYPOINT [ "/p2p-chat" ]
ENTRYPOINT [ "/docker-entrypoint.sh" ]
