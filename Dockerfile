FROM golang:1.9-alpine
COPY . /go/src/github.com/ahmetb/freshpod
WORKDIR /go/src/github.com/ahmetb/freshpod
RUN go install .

FROM alpine:3.7
COPY --from=0 /go/bin/freshpod /freshpod
ENTRYPOINT ["/freshpod"]
