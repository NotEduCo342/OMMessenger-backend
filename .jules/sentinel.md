## 2024-06-26 - Stored XSS via File Upload and Predictable Filenames

**Vulnerability:**
The `UploadAttachment` handler allowed uploading files with any `Content-Type` and used a predictable, timestamp-based filename format (`strconv.FormatInt(time.Now().UnixNano(), 10)`). This could allow an attacker to upload a malicious HTML file (or SVG with embedded scripts) and accurately predict its URL to execute a Stored Cross-Site Scripting (XSS) attack against other users. Additionally, when serving files, the `GetMedia` handler lacked basic defense-in-depth security headers like `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; sandbox`.

**Learning:**
Relying solely on the original file extension or missing a strict `Content-Type` validation enables attackers to host executable content on our domain. Timestamp-based names expose the application to file enumeration and race conditions (unintended overwrites). Missing security headers when downloading files means browsers might attempt to "sniff" the MIME type and execute the file if the `Content-Type` header is spoofed.

**Prevention:**
Always enforce a strict whitelist for `Content-Type` (e.g., `image/jpeg`, `image/png`, `video/mp4`) *before* processing an upload. Generate filenames using cryptographically secure UUIDs (`uuid.NewString()`) instead of predictable values. Always apply defense-in-depth headers like `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; sandbox` on file download/serving endpoints to prevent browsers from executing the content as active HTML/JS context.
