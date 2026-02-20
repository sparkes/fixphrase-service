# FixPhrase Service

A minimal, self-contained Go webservice implementing the **FixPhrase** coordinate encoding algorithm.

FixPhrase converts latitude/longitude pairs into a deterministic four-word phrase and can decode phrases back into approximate coordinates.

This repository is open to contributions, experimentation, and extension. If you're interested in geospatial encoding systems, lightweight Go services, or containerised API deployments, you're welcome here.

## About FixPhrase

> FixPhrase is a simple algorithm and word list that enable converting a pair of GPS coordinates to a four-word phrase and back again. This could be useful in situations where it's easier to deal with a few words instead of a long series of numbers.

[FixPhrase Project Repository](https://source.netsyms.com/Netsyms/fixphrase.com)

[FixPhrase Demo Site](https://fixphrase.com/)

---

# ⚠️ Project Status

**Pre-1.0.0 (Early Release)**

This project is currently in active development. Until version `1.0.0` is released:

- APIs may change
- Configuration may evolve
- Internal refactors may occur
- Backwards compatibility is not guaranteed

### Recommended Usage (Pre‑1.0.0)

This service is best used in:

- Local development environments
- Internal tooling
- Lab/testing deployments
- Behind a reverse proxy with authentication

It is **not recommended** to expose this service directly to the public internet without:

- TLS termination
- Authentication
- Rate limiting
- Monitoring
- Reverse proxy (NGINX, Caddy, Traefik, etc.)

Once 1.0.0 is reached, public deployment guidance will be formally documented.

---

# What This Project Provides

- Embedded official `wordlist.json` ([list and licence](https://github.com/sparkes/fixphrase-service/tree/main/wordlist) )
- Deterministic FixPhrase encode/decode implementation
- REST API (GET + POST)
- Graceful shutdown (SIGINT/SIGTERM aware)
- Version stamping (version, commit, build date)
- Cross-platform release binaries
- Multi-architecture Docker images (amd64 + arm64)
- GitHub Release automation

---

# How FixPhrase Works

Coordinates are:

1. Rounded to 4 decimal places
2. Shifted to remove negative ranges
3. Converted into digit blocks
4. Mapped into four word indices using offset ranges

Accuracy levels:

| Words | Approximate Accuracy |
|--------|---------------------|
| 2      | ~11 km |
| 3      | ~1.1 km |
| 4      | ~11 meters |

---

# API

## Health Check

```
GET /healthz
```

Example response:

```json
{
  "ok": true,
  "service": "FixPhrase Service",
  "version": "v0.1.0",
  "commit": "abc123",
  "buildDate": "2026-02-19T12:00:00Z"
}
```

---

## Encode Coordinates

### GET

```
GET /encode?lat=52.5902&lon=-2.1304
```

### POST

```
POST /encode
Content-Type: application/json

{
  "lat": 52.5902,
  "lon": -2.1304
}
```

Example response:

```json
{
  "lat": 52.5902,
  "lon": -2.1304,
  "phrase": "croak lusty subwoofer trustful",
  "words": [
    "croak",
    "lusty",
    "subwoofer",
    "trustful"
  ]
}
```

---

## Decode Phrase

### GET

```
GET /decode?phrase=croak%20lusty%20subwoofer%20trustful
```

### POST

```
POST /decode
Content-Type: application/json

{
  "phrase": "croak lusty subwoofer trustful"
}
```

Example response:

```json
{
  "inputWordsSorted": [
    "croak",
    "lusty",
    "subwoofer",
    "trustful"
  ],
  "result": {
    "inputWords": [
      "croak",
      "lusty",
      "subwoofer",
      "trustful"
    ],
    "canonicalPhrase": "croak lusty subwoofer trustful",
    "lat": 52.5902,
    "lon": -2.1304,
    "accuracyDegrees": 0.0001
  }
}
```

---

# Configuration

Configuration may be provided via environment variables or an optional `.env` file.

## Example `.env`

```
ADDR=:7080
SERVER_NAME=FixPhrase Service
REPO_URL=https://github.com/youruser/yourrepo
GHCR_IMAGE=ghcr.io/youruser/yourrepo
```

## Environment Variables

| Variable | Default | Description |
|-----------|----------|------------|
| `ADDR` | `:7080` | Bind address and port |
| `SERVER_NAME` | `fixphrase` | Service name shown in health output |
| `REPO_URL` | empty | Repository metadata |
| `GHCR_IMAGE` | empty | Docker image metadata |

Environment variables override `.env` values.

---

# Running

## Local Development

```
go run .
```

## Build Binary

```
go build -o fixphrase
./fixphrase
```

---

# Docker

## Run Released Image

```
docker run -p 7080:7080 ghcr.io/<owner>/<repo>:v0.1.0
```

## With Custom Environment

```
docker run \
  --env-file .env \
  -p 9090:9090 \
  ghcr.io/<owner>/<repo>:v0.1.0
```

---

# Contributing

Contributions are welcome.

[see contributing.md for details](https://github.com/sparkes/fixphrase-service/blob/main/docs/contributing.md)

---

# Community Code of Conduct

Starting as we mean to go on.

Be Nice, No Nazi's, No Terf's, No Arseholes

[Code of Conduct](https://github.com/sparkes/fixphrase-service/blob/main/docs/code_of_conduct.md)

---

# Security Policy

If you see it and can't fix it report it.  This is how we keep the internet safe for all of us. 

[Security Policy](https://github.com/sparkes/fixphrase-service/blob/main/docs/security.md)

---

# Roadmap Toward 1.0.0

No Promises ;)

- Structured logging
- Request logging middleware
- Optional authentication
- Rate limiting middleware
- Metrics endpoint
- API versioning
- Input validation hardening
- Public deployment guidance

---

# License

MIT License

Copyright (c) 2026 Stephen Parkes and FixPhrase-Service Contributors 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

