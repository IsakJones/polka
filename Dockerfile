FROM golang:1.17 AS polka

WORKDIR /polka

# arg[0] -> local, arg[1] -> in the image
COPY . .

RUN make clean
RUN make docker_prep

FROM polka AS balancer
WORKDIR /polka/balancer
CMD ./bin/polkabalancer

FROM polka AS cache
WORKDIR /polka/cache
CMD ./bin/polkacache

FROM polka AS receiver
WORKDIR /polka/receiver/node0
CMD ./polkareceiver

FROM polka AS settler
WORKDIR /polka/settler
CMD ./bin/polkasettler

FROM polka AS generator
WORKDIR /polka/generator
CMD ./bin/polkagenerator -w=1000 -t=1000