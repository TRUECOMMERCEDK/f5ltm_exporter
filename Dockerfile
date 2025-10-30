FROM golang:1.24 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download -x

COPY . ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.release=`git rev-parse --short=8 HEAD`'" -o /bin/server ./cmd/f5ltm_exporter

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /bin/server ./
ENTRYPOINT ["./server"]
