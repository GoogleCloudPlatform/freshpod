FROM golang:1.9-alpine
RUN apk add --no-cache git
COPY . /go/src/github.com/ahmetb/killa
WORKDIR /go/src/github.com/ahmetb/killa
# TODO(ahmetb) we shouldn't neet "go get" when client-go works with dep.
RUN go get ./...
RUN go install .

FROM alpine
COPY --from=0 /go/bin/killa /killa
ENTRYPOINT ["/killa"]
