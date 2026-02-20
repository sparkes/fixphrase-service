# Security Policy

## Supported Versions

This project is currently in pre-1.0.0 development.

Security fixes will be applied to the latest release series only.

| Version | Supported |
|---------|-----------|
| < 1.0.0 | ⚠️ Best effort |
| 1.x.x   | ✅ Yes (once released) |

Until version 1.0.0, backward compatibility is not guaranteed and security hardening is ongoing.

---

## Reporting a Vulnerability

If you discover a security vulnerability, please do **not** open a public issue.

Instead:

- Contact the maintainer privately
- Provide a clear description of the issue
- Include reproduction steps if possible
- Include impact assessment if known

We will:

1. Acknowledge receipt within a reasonable timeframe
2. Investigate and reproduce the issue
3. Provide a fix or mitigation plan
4. Coordinate disclosure if appropriate

---

## Scope

This service is intentionally minimal and stateless. Security considerations include:

- Input validation and parsing
- HTTP request handling
- Resource exhaustion risks
- Denial-of-service exposure
- Deployment configuration

Because this service is pre-1.0.0, it is strongly recommended that it be deployed behind:

- A reverse proxy
- TLS termination
- Authentication
- Rate limiting

The project does not currently include built-in authentication, rate limiting, or request throttling.

---

## Responsible Disclosure

We appreciate responsible disclosure and will work cooperatively with reporters to resolve issues.

Public disclosure should occur only after a fix or mitigation has been prepared.

---

Thank you for helping keep the project secure.

