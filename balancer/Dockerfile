FROM golang:1.17

WORKDIR /balancer

# arg[0] -> local, arg[1] -> in the image
COPY . .
COPY . ../go.mod
COPY . ../go.sum

# Build binary
RUN go build -o bin/balancer ./src

# Binary should already be in bin/
CMD [ "./bin/balancer" ]
