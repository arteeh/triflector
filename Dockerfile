FROM golang AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /triflector

FROM gcr.io/distroless/base-debian12 AS run
WORKDIR /
COPY --from=build /triflector /triflector
USER nonroot:nonroot
EXPOSE 3334
ENTRYPOINT ["/triflector"]