# syntax=docker/dockerfile:1

FROM gcr.io/distroless/static:nonroot
WORKDIR /

# GoReleaser dockers_v2 puts artifacts under $TARGETPLATFORM/
ARG TARGETPLATFORM

# Optional port metadata (service still binds via ADDR env at runtime)
ARG PORT=7080
ENV PORT=${PORT}

COPY $TARGETPLATFORM/fixphrase /fixphrase

EXPOSE ${PORT}
USER nonroot:nonroot
ENTRYPOINT ["/fixphrase"]
