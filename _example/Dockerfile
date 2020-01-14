FROM golang:latest as build
WORKDIR /go/src/autocache
ADD . /go/src/autocache
RUN go get -d -v ./...
RUN go build -o /go/bin/autocache main.go

FROM gcr.io/distroless/base
COPY --from=build /go/bin/autocache /
ENTRYPOINT [ "/autocache" ]