FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o kopdsync

FROM scratch

COPY --from=build /app/kopdsync /usr/bin/kopdsync

EXPOSE 8080

ENTRYPOINT ["/usr/bin/kopdsync"]
