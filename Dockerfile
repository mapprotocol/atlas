# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build atlas in a stock Go builder container
FROM golang:1.18 as builder

RUN apt install make  gcc  git

ADD . /atlas
RUN cd /atlas && make atlas

# Pull atlas into a second stage deploy ubuntu container
FROM ubuntu:latest

COPY --from=builder /atlas/build/bin/atlas /usr/local/bin/
RUN chmod +x /usr/local/bin/atlas

EXPOSE 7445 20101 20101/udp
ENTRYPOINT ["atlas"]

# Add some metadata labels to help programatic image consumption
LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"

