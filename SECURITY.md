# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in ossf-scout, please report it through [GitHub's private vulnerability reporting](https://github.com/nagyonmarci/ossf-scout/security/advisories/new).

Do not open a public issue for security vulnerabilities.

We will acknowledge the report within 48 hours and aim to release a fix within 14 days for confirmed issues.

## Scope

The following are considered in scope:

- Remote code execution or privilege escalation via the web server
- Injection vulnerabilities (SQL, command, etc.)
- Authentication/authorization bypass
- Information disclosure of sensitive data

The following are **out of scope**:

- Vulnerabilities in dependencies (report those upstream)
- Issues that require physical access to the machine
- Self-XSS or issues requiring the attacker to already have admin access

## Supported Versions

Only the latest version on the `main` branch is actively maintained.
