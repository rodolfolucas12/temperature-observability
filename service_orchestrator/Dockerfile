FROM golang:1.23 as build
WORKDIR /app
COPY ../go.mod ../go.sum ./
RUN go mod download

COPY . ./
WORKDIR /app/serviceb
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/serviceb

FROM scratch
WORKDIR /app
COPY --from=build /app/serviceb .
ENTRYPOINT ["/app/serviceb"]