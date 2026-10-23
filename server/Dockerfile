FROM golang:1.25-alpine AS development

WORKDIR /app

RUN go install github.com/air-verse/air@v1.63.0

COPY go.mod ./
RUN go mod download
COPY . .

EXPOSE 8090

CMD ["air", "-c", ".air.toml"]

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:3.21

RUN adduser -D -H -u 10001 appuser
USER appuser

COPY --from=build /server /server

EXPOSE 8090

ENTRYPOINT ["/server"]
