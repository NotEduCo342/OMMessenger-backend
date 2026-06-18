
## 2024-06-18 - [Fix Stored XSS via File Upload]
**Vulnerability:** Found Stored XSS vulnerability in file upload endpoints where any file type could be uploaded without content type validation, and served without defense-in-depth security headers like `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; sandbox`.
**Learning:** In fiber handlers handling media upload/download via S3 or local storage, file type validation must be strictly whitelisted during upload, and security headers must be enforced during download to mitigate Stored XSS attacks.
**Prevention:** Always whitelist `Content-Type` in file upload handlers (e.g. `UploadAttachment`) and apply strict headers (`X-Content-Type-Options: nosniff`, `Content-Security-Policy: default-src 'none'; sandbox`) to responses in media download endpoints (e.g. `GetMedia`).
