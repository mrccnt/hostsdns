FROM alpine:latest
COPY dist/hostsdns /app/hostsdns
COPY config.json /app/config.json
WORKDIR /app
EXPOSE 53/udp
ENTRYPOINT ["/app/hostsdns", "-f", "/app/config.json"]
