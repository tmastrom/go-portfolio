FROM golang:1.25 AS build-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY /cmd/main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /

COPY --from=build-stage /out /out
COPY /public/ /public/
COPY /css/ /css/
COPY /posts/ /posts/
COPY /templates/ /templates/

USER nonroot:nonroot
ENTRYPOINT ["/out"]