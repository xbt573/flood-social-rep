FROM golang AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download && go mod verify

COPY . .

# Building CGO statically for sqlite3 module
RUN CGO_ENABLED=1 go build -ldflags="-extldflags=-static" -tags sqlite_omit_load_extension

FROM alpine

COPY --from=build /app/flood-social-rep /
CMD ["/flood-social-rep"]
