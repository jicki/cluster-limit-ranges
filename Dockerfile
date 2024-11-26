## Build Golang
FROM golang:1.23 AS builder

#ENV GOPROXY https://goproxy.cn
#ENV GOSUMDB sum.golang.google.cn
COPY . .

RUN set -ex \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o cluster-limit-controller .

FROM alpine:3.15 AS final

ARG TZ="Asia/Shanghai"

ENV TZ=${TZ}
ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8
ENV LANGUAGE=en_US:en

RUN apk update && apk upgrade

RUN set -ex \
    && apk add bash tzdata \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/cluster-limit-controller /app/

RUN chmod +x /app/cluster-limit-controller

EXPOSE 8888

ENTRYPOINT ["/app/cluster-limit-controller"]

