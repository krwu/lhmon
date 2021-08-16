FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /lhmon

FROM alpine:latest

WORKDIR /app

COPY --from=build /lhmon /app/lhmon

CMD ["/app/lhmon"]
