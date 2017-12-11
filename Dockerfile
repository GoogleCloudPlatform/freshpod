FROM golang:1.9-alpine
RUN apk add --no-cache git
COPY . /go/src/github.com/ahmetb/freshpod
WORKDIR /go/src/github.com/ahmetb/freshpod
RUN go install .

FROM alpine
COPY --from=0 /go/bin/freshpod /freshpod
ENTRYPOINT ["/freshpod"]
