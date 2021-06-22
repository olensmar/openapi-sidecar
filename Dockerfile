# Use the Go v1.16 image for the base.
FROM golang:1.16

RUN mkdir /main
WORKDIR /main
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY main.go .

# Run the proxy on container startup.
ENTRYPOINT [  "go" ]
CMD [ "run", "main.go" ]
