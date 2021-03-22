FROM alpine:3.7

COPY ./entrypoint.sh /
COPY ./bin /usr/local/bin
RUN chmod +x entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["log-exploration-api"]
