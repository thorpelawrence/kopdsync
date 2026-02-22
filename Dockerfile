FROM golang:1.26-alpine AS build

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN go build -o kopdsync

FROM scratch

COPY --from=build /app/kopdsync /usr/bin/kopdsync

EXPOSE 8080

ENTRYPOINT ["/usr/bin/kopdsync"]
