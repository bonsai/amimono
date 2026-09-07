FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/amimono-api ./cmd/amimono-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/amimono-api /amimono-api
EXPOSE 8080
ENTRYPOINT ["/amimono-api"]
