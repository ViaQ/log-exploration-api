FROM alpine:3.7

COPY ./entrypoint.sh /
COPY ./bin /usr/local/bin

ENTRYPOINT ["/entrypoint.sh"]
CMD ["log-exploration-api"]
