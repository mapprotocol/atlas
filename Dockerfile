# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build atlas in a stock Go builder container
FROM golang:1.18-alpine as builder

RUN apk add --no-cache gcc musl-dev linux-headers git make

ADD . /atlas
RUN cd /atlas && make atlas

# Pull atlas into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /atlas/build/bin/atlas /usr/local/bin/

EXPOSE 7445 30303 30303/udp
ENTRYPOINT ["atlas"]

# Add some metadata labels to help programatic image consumption
LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
