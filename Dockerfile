FROM golang:1.17

WORKDIR /polka

# arg[0] -> local, arg[1] -> in the image
COPY . .

# build binaries
RUN go build -o build/payments-api ./apis/src/ 
RUN go build -o build/generator ./generator
RUN go build -o build/balancer ./balancer
RUN go build -o build/cache ./cache

WORKDIR /polka/apis/api0

# docker run -p 8081:8080 payments:0.17 /polka/build/payments-api

# FROM alpine

# WORKDIR /app

# COPY --from=builder /polka/build/cache cache
# COPY --from=builder /polka/build/generator generator
# COPY --from=builder /polka/build/balancer balancer
# COPY --from=builder /polka/build/payments-api payments-api

# CMD cache --port 8081


# COPY
# WORKDIR
# RUN
# CMD

# docker run -it golang:1.17 sh
# docker build -t payments:0.1 .
