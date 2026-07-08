FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/backend
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/notifications ./cmd/notifications

FROM alpine:3.22

RUN addgroup -S homelytics && adduser -S -G homelytics homelytics
WORKDIR /app
COPY --from=build /out/backend /app/backend
COPY --from=build /out/notifications /app/notifications
USER homelytics

EXPOSE 8080
CMD ["/app/backend"]
