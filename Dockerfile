FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY out .
CMD ["./app"]