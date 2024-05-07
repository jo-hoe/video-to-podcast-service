FROM golang:1.22-alpine as build

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM jrottenberg/ffmpeg:4.1-vaapi

COPY --from=build /go/bin/app /

ENTRYPOINT ["/app"]