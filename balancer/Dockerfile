FROM golang:1.17

WORKDIR /balancer

# arg[0] -> local, arg[1] -> in the image
COPY . .

# Binary should already be in bin/
CMD [ "./bin/balancer" ]
