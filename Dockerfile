# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/restaurante-back .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/restaurante-back /app/restaurante-back
COPY conf ./conf
COPY static ./static
USER app
ENV BEEGO_RUNMODE=prod TZ=America/Bogota
EXPOSE 8080
CMD ["/app/restaurante-back"]
