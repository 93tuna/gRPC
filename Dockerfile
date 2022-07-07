# final stage
FROM ubuntu:20.04

RUN apt-get update;apt-get install -y ca-certificates;

COPY ./dist/ /app/

RUN cd /app/

ENTRYPOINT ["/app/grpc"]