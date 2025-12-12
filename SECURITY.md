# Security Policy

## Security Overview

This password generator (`pw`) is designed with security as a primary concern. This document outlines the security features, considerations, and best practices.

## Security Features

### 1. **Cryptographic Security**
- Uses **HMAC-SHA256** for password generation, a cryptographically secure hash function
- Deterministic generation ensures the same password for the same site/secret combination
- No randomness means passwords are reproducible from the master secret alone

### 2. **Secret Protection**
The application enforces strict security measures for the secret file:
- **File permissions**: Must be exactly `0600` (read/write for owner only)
- **Ownership validation**: File must be owned by the current user
- **Location**: `~/.config/pw/secret` (user-specific, not shared)

If these conditions are not met, the application will refuse to run.

### 3. **Path Traversal Protection**
Site names are sanitized before use:
- Forward slashes (`/`) are replaced with underscores (`_`)
- Spaces are replaced with underscores (`_`)
- Names are normalized to lowercase

This prevents path traversal attacks when loading site-specific policies.

### 4. **Debug Mode Safety**
Debug logging has been carefully designed to avoid leaking sensitive information:
- Site names are **not** logged
- Password characters (original or modified) are **not** logged
- Only metadata about policy application is logged (boolean flags, positions)

**Important**: Even with these protections, debug mode should only be used in secure, private environments.

## Security Considerations

### Legacy Mode - DEPRECATED AND INSECURE
⚠️ **WARNING**: The `-legacy` flag is deprecated and should **NOT** be used.

The legacy mode currently returns a static dummy password and is maintained only for backward compatibility during migration. Using this mode is a **critical security vulnerability**.

**If you are using legacy mode**, please:
1. Migrate to the standard password generation mode immediately
2. Change all passwords that were generated using legacy mode
3. Do not use the `-legacy` flag in production

### Secret Generation Best Practices

When creating your secret key file:

```sh
# Generate a high-entropy secret (recommended)
head -c 64 /dev/urandom | base64 > ~/.config/pw/secret
chmod 0600 ~/.config/pw/secret
```

**Do NOT**:
- Use simple passwords or phrases as your secret
- Share your secret file with anyone
- Store your secret in version control
- Use the same secret across multiple machines (if one is compromised)
- Copy your secret over insecure channels

### Site Policy Files

Site policy files (in `~/.config/pw/sites/`) control password requirements:
- These files are less sensitive than the secret file
- Permissions are not strictly enforced for policy files
- However, restricting access is still recommended

### Password Backup Strategy

Since passwords are deterministically generated:
- **Backup your secret file securely** - this is your master key
- Store the backup in a secure location (encrypted drive, password manager, etc.)
- Consider printing and storing in a physical safe as ultimate backup
- **DO NOT** store passwords themselves - regenerate them from the secret

## Reporting Security Issues

If you discover a security vulnerability in this project, please report it by:

1. **Do NOT** open a public GitHub issue
2. Email the maintainer directly with details
3. Include steps to reproduce if possible
4. Allow time for a fix before public disclosure

## Security Audit History

### December 2025 - Security Review
- **Fixed**: Debug logging information leakage (removed site names and password characters from logs)
- **Fixed**: Added deprecation warnings for insecure legacy mode
- **Verified**: HMAC-SHA256 implementation is correct
- **Verified**: Secret file permission checks are properly enforced
- **Verified**: Path traversal protection is working correctly
- **Verified**: No vulnerabilities found by CodeQL static analysis

## Best Practices for Users

1. **Generate a strong secret**: Use `/dev/urandom` or equivalent to generate high-entropy secrets
2. **Protect your secret file**: Ensure proper permissions (0600) and ownership
3. **Avoid debug mode**: Only use `-debug` flag when troubleshooting in a secure environment
4. **Regular security updates**: Keep the tool updated to get security fixes
5. **Site name consistency**: Use the same site name format consistently (e.g., always lowercase domain names)
6. **Policy management**: Review and update site policies as password requirements change

## Security Limitations

Users should be aware of these inherent limitations:

1. **Master secret is single point of failure**: If your secret file is compromised, all passwords are compromised
2. **Deterministic generation**: Same inputs always produce same outputs (this is by design, but means no per-password salt)
3. **No password rotation reminders**: The tool generates the same password each time unless you change your secret
4. **Terminal history**: Site names may be visible in shell history if you echo them into the tool
5. **Process listings**: Site names may be briefly visible in process listings

## Recommended Security Enhancements for Users

For additional security:
- Use a separate secret for high-value accounts
- Periodically rotate your secret (requires changing all passwords)
- Run the tool in a secure, isolated environment
- Consider using `history -d` to remove sensitive commands from shell history
- Use input redirection from files instead of echo to avoid process listing exposure

## License

This security policy is part of the pw project and follows the same license terms.
