# Build tools exist only in this stage. Refresh this pinned version for Go security updates.
FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY go.mod main.go error.html ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o /error-pages .

FROM scratch
COPY --from=build --chown=65532:65532 /error-pages /error-pages
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/error-pages"]
