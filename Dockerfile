FROM golang as builder
WORKDIR /go/src/github.com/dmage/boater
ADD . .
RUN CGO_ENABLED=0 go install .

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/boater /usr/bin/boater
USER 1000
