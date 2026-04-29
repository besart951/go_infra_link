Security Improvements

Fix dependency advisories first. pnpm audit found 24 vulnerabilities: 7 high, 12 moderate, 5 low. Key packages: xlsx, vite, rollup, @sveltejs/kit, svelte, devalue, postcss, picomatch.
Replace or isolate xlsx. The audit reports high-severity SheetJS advisories with no npm patched version. Prefer backend-side Excel parsing/export through Go excelize, or sandbox the worker and strictly validate files.
Update Go dependency chain. govulncheck reports GO-2025-4233 in github.com/quic-go/quic-go@v0.54.0; fixed in v0.57.0.
Disable public production source maps. vite.config.ts (line 71) has sourcemap: true; the build emits 158 .map files totaling 6.91 MB.
Add a CSP header in Caddyfile (line 1). You already have good base headers, but no Content-Security-Policy.
Configure Gin trusted proxies explicitly. router.go (line 17) uses gin.Default() without SetTrustedProxies; that matters because login rate limiting uses ClientIP().
Performance Improvements

Optimize field-device list queries. field_device_repo.go (line 312) preloads deep relations and BACnet objects for paginated lists. Use a lightweight list DTO and load BACnet/specification details only on detail/edit screens.
Reduce the max list size for filtered field-device queries. field_device_repo.go (line 243) allows up to 1000, which will hurt as data grows.
Cache metadata endpoints. /facility/field-devices/options is already ~205 ms on seed data; add ETags or server-side cache invalidated by apparat/system-part/object-data changes.
Shrink frontend payloads. The Excel route emits the largest client node at ~369 KB plus a ~329 KB Excel worker. Keep Excel code strictly route/worker lazy-loaded and avoid shared imports from normal app routes.
Add long-lived cache headers for hashed assets under /_app/immutable/* in Caddy. That is low risk because filenames are content-hashed.