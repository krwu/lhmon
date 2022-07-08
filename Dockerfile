FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /lhmon

FROM alpine:latest

ENV TZ Asia/Shanghai

RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && apk del tzdata

WORKDIR /app

COPY --from=build /lhmon /app/lhmon

CMD ["/app/lhmon"]
