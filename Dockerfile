FROM golang:alpine as build

WORKDIR $GOPATH/src/github.com/stevenctl/grpc-xds-sample
COPY . .
RUN go mod download
RUN go build -o /go/bin/app .

FROM alpine
COPY --from=build /go/bin/app /
CMD ["/app"]
