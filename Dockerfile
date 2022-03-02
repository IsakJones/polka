FROM golang:1.17 AS polka

WORKDIR /polka

# arg[0] -> local, arg[1] -> in the image
# COPY go.mod .
# COPY go.sum .
# RUN go get

COPY . .

RUN ./scripts/clean_up.sh 
RUN ./scripts/prep.sh 2
# RUN go build -o generator/bin/polkagenerator ./generator/src/
# RUN go build -o balancer/bin/polkabalancer ./balancer/src/
# RUN go build -o receiver/bin/polkareceiver ./receiver/src/ 
# RUN go build -o cache/bin/polkacache ./cache/src/

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