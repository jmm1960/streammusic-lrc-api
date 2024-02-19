FROM golang:1.20 as builder

ADD . /ylyric
RUN cd /ylyric && CGO_ENABLED=0 go build -o ylyric main.go

FROM alpine
COPY --from=builder /ylyric/ylyric /usr/local/bin/
EXPOSE 8092
CMD ylyric -ncmapi http://neteasecloudmusicapi:3000