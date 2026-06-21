## 2024-05-19 - Fast Exists Query in GORM
**Learning:** Using `Count(&count)` in GORM for an existence check evaluates the entire table/index, whereas `.Select("1").Limit(1).Find(&dummy)` will fail fast and return instantly once a match is found, improving performance on large tables.
**Action:** When evaluating whether a row exists or a condition is true, utilize the `.Select("1").Limit(1).Find(&dummy)` pattern over `Count()`.

## 2025-02-12 - Prevent Stored XSS in Media Uploads
**Vulnerability:** Serving user uploaded media with a non-restricted Content-Type and without `nosniff` and CSP headers could potentially lead to Stored XSS if malicious scripts are uploaded and served with e.g., `text/html`.
**Learning:** Fiber static file serving or streaming needs explicit restrictive `Content-Security-Policy` with `sandbox` and `X-Content-Type-Options: nosniff` to neutralize content spoofing and execution. Also use UUID instead of timestamp to avoid predictable URLs.
**Prevention:** Always enforce strict `Content-Type` whitelists upon upload, generate cryptographically random filenames (UUID), and serve files with strict CSP sandbox headers.
