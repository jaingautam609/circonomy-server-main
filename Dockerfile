FROM golang:1.19 AS build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /server

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/circonomy-server cmd/main.go

FROM alpine

WORKDIR /

COPY --from=build /server/bin .
COPY --from=build /server/database/migrations ./database/migrations

EXPOSE 5000
ENTRYPOINT ["./circonomy-server"]
