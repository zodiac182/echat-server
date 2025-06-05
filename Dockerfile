FROM docker.m.daocloud.io/library/golang:alpine AS builder

WORKDIR /go/src/echat/
COPY ./ .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go mod tidy \
    && go build -o echat-server .

FROM docker.m.daocloud.io/library/alpine:3.22

ENV LANG="C.UTF-8"
ENV TZ="Asia/Shanghai"

RUN echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.22/main/" > /etc/apk/repositories \
    && echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.22/community/" >> /etc/apk/repositories \
    && echo "http://mirrors.ustc.edu.cn/alpine/v3.22/main/" >> /etc/apk/repositories \
    && echo "http://mirrors.ustc.edu.cn/alpine/v3.22/community/" >> /etc/apk/repositories \
    && echo "http://mirrors.nju.edu.cn/alpine/v3.22/main/" >> /etc/apk/repositories \
    && echo "http://mirrors.nju.edu.cn/alpine/v3.22/community/" >> /etc/apk/repositories \
    && apk add -U tzdata \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /echat

COPY --from=builder /go/src/echat/echat-server ./

# 设置应用的运行级别
# ENV LOG_LEVEL=info
ENV APP_MODE=production
ENV GIN_MODE=release


ENTRYPOINT ["/echat/echat-server"]
