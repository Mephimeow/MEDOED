FROM alpine:latest AS builder 
RUN apk add --no-cache g++ cmake make curl-dev linux-headers
WORKDIR /app
COPY . .
RUN mkdir build && cd build && cmake .. && make

FROM alpine:latest
RUN apk --no-cache add libcurl 
COPY --from=builder /app/build/agent /usr/local/bin/agent
CMD ["agent"]