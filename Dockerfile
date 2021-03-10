FROM golang:alpine

ENV GIN_MODE=release
ENV PORT=8080

WORKDIR /usr/local/go/src/logexplorationapi/pkg

COPY pkg /usr/local/go/src/logexplorationapi/pkg

RUN go build main.go

EXPOSE $PORT

ENTRYPOINT ["./main"]