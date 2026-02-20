# Contributing to FixPhrase Service

Thank you for your interest in contributing.

This project aims to remain:

- Minimal
- Deterministic
- Easy to deploy
- Easy to reason about

Before submitting changes, please read the guidelines below.

---

## Development Setup

Requirements:

- Go 1.22+
- Docker (optional)

Clone your fork:

```
git clone https://github.com/<your-user>/fixphrase-service.git
cd fixphrase-service
```

Run locally:

```
go run .
```

---

## Contribution Types

We welcome:

- Documentation improvements
- Bug fixes
- Tests
- Middleware enhancements (logging, metrics, rate limiting)
- Performance improvements
- Security hardening

Please open an issue before large architectural changes.

---

## Pull Request Guidelines

- Keep PRs focused and small where possible
- Include clear commit messages
- Update documentation when relevant
- Avoid introducing breaking changes without discussion

If your change affects API behavior, please document it clearly.

---

## Release Policy

- Breaking changes are allowed before 1.0.0
- After 1.0.0, semantic versioning will be followed

---

Thank you for helping improve the project.

