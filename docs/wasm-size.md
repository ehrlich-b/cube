# Pages wasm size measurements

Baseline: `9ed6d04`, Go 1.26.2, `GOOS=js GOARCH=wasm go build -p 2 -trimpath`.
All sizes below use decimal MB; gzip level 9 and Brotli quality 11 are per-file
estimates, not an assumption about GitHub Pages response headers.

| Artifact | Before raw / gzip / Brotli MB | After raw / gzip / Brotli MB |
| --- | --- | --- |
| Initial runtime graph | 11.273 / 6.548 / 6.066 | 4.701 / 1.386 / 1.054 |
| Wasm alone | 11.172 / 6.518 / 6.040 | 4.597 / 1.355 / 1.027 |
| Lazy assets, all dimensions | 0 (embedded above) | 4.544 / 4.500 / 4.474 |
| Full export | 11.273 / 6.548 / 6.066 | 9.245 / 5.886 / 5.528 |

Lazy 3x3 coordinates cost 3.011 / 2.982 / 2.959 MB. The 4x4, 5x5, 6x6 and
7x7 assets cost respectively 0.081 / 0.081 / 0.081, 0.630 / 0.622 / 0.622,
0.315 / 0.312 / 0.312, and 0.507 / 0.502 / 0.501 MB. These files already contain
gzip data; the HTTP compression estimates describe additional per-file compression.
Without Binaryen, initial bytes are 5.049 / 1.395 / 1.049 MB. Its optional pass
saves 348,079 raw bytes and roughly 9 KB gzip while costing roughly 4 KB Brotli.

Three paired runs in fresh headless Chromium 153 profiles, alternating order,
gave these medians (milliseconds):

| Timing | Before | After |
| --- | ---: | ---: |
| Initial load, 390×844, 4× CPU throttle | 684 | 563 |
| First 3x3 solve of `R`, 1× CPU | 1,618 | 1,238 |
| First 7x7 solve, including fetch and playback preparation | 2,602 | 2,603 |
| 7x7 worker completion, including fetch and initialization | 1,122 | 1,074 |
| 7x7 wasm solve only | 885 | 663 |

The 7x7 input is `2R U 2F' Rw D2 L B'`, with all 231 returned moves independently
replayed by the browser harness. Every trial fetches coordinates and the 7x7
asset in its fresh worker. Transfer uses in-memory Playwright routing without
network throttling, so these timings do not estimate cellular download latency.
The full smoke test additionally exercises random scrambles on every size and
checks responsiveness and cancellation on 7x7.

The baseline wasm is 11,172,246 bytes. Its data section is 6,956,779 bytes, code
is 4,088,746 bytes, and function names are 115,559 bytes. Embedded assets:

| Asset | Bytes |
| --- | ---: |
| 3x3 coordinates | 3,337,787 |
| 4x4 reduction | 81,369 |
| 5x5 reduction | 629,581 |
| 6x6 reduction | 315,493 |
| 7x7 reduction | 506,727 |
| Algorithm JSON | 74,002 |

`go tool nm -size -sort size .scratch/baseline.wasm` reports “unrecognized object
file” on this toolchain. `tools/inspect-wasm.mjs` reads the wasm code and name
sections instead. Largest named baseline functions are Unicode map
initialization (50,636 bytes), API dispatch (29,270), runtime metric setup
(29,218), JSON literal decoding (26,719), and CFOP solving (26,405).
The runtime's free functions account for 739,383 code bytes; gob adds over
230 KB across its free functions and encoder/decoder methods. Regex parsing
and reflective formatting also appear prominently in the function report.

The browser now loads a checked binary coordinate asset (3,010,837 bytes)
containing every native coordinate. A native test compares all decoded fields
and the reconstructed frontier map with the original table, then checks
truncation, expiration and re-signed invalid schemas. Native gob caching and
embedded assets retain their original format and behavior.

Tables load when Kociemba, pattern search or reduction needs them. Reduction
fetches only its selected dimension. Fetch SRI and a second SHA-256/size check
precede installation; existing table dimension, move fingerprint and gzip
checks precede use. The Pages test verifies no table request on initial load,
independent replay of the first 7x7 result, and retry after missing/corrupt data.
The in-memory smoke transport serves the same files without an external API.

The build strips symbols, omits native gob/filesystem coordinate caching and
the CLI-only large-table option, and replaces regex tokenization with scanners
tested against the original regex grammars. Optional installed Binaryen runs
`-Oz` with only features already used by Go. `WASM_OPT=off` keeps the build
portable and has its own enforced size budget. Solve/search workers reuse
compiled code with independent Go heaps. A separate NxN wasm was considered;
it would duplicate the Go runtime and require another compiled module download.
Keeping one shared module and loading data avoids that extra transfer.

`make test-pages` passed with and without Binaryen. The final optimized build
also passed `make test-web`,
`node web/test/smoke.mjs --in-memory`, `go test -p 2 ./...`, CLI E2E 138/138,
`make test-nxn` (200 independently replayed fresh-process cases per tested size),
`make test-docs` (30/30 runnable blocks), and `go vet -p 2 ./...`. CLI and all
`./tools/...` programs were built into `dist/` before E2E. No remote writes,
dependency installation, or changes to another checkout were performed.

The raw initial-load goal of 4 MB remains unmet; the compressed-load goal of
2 MB is met with substantial room. The remaining wasm largely consists of Go
runtime, JSON/reflection support, and solver code rather than embedded tables.
No solver, search depth, latency oracle or replay check was relaxed.

Reproduce the measurements with `make test-pages` and inspect the generated
`.scratch/pages-readiness.json`. For named attribution, build a measurement
wasm without `-s`, then run `node tools/inspect-wasm.mjs <measurement.wasm>`.
Run builds and scripts with the repository's required background priority.

## Deterministic two-phase / native 2x2 merge

The merged `73219d8` optimized wasm grew from `b67ddaa`'s 4,597,000 bytes
to 4,734,489 bytes. The native 2x2 fast cache imported `encoding/gob` in
the wasm build even though its entry point returned immediately on wasm.
Go removed the unused cache functions but retained gob package initialization,
its type registry and their dependencies. A named build contained 93 gob
functions totaling 71,555 code bytes, including `init`, `newTypeObject`,
`buildTypeInfo` and `registerBasics`.

Build constraints now exclude the cache implementation and its filesystem
tests on every wasm target; a wasm stub preserves the original no-op entry
point. Native cache loading, validation, saving and fallback remain unchanged.
No gob functions remain in the named browser build. This removes 133,523
optimized bytes without changing the deterministic search or its complete
fallback. The remaining 3,966 bytes above `b67ddaa` reflect the merged solver
changes and binary layout.

Section payload sizes from `tools/inspect-wasm.mjs` (optimized builds, bytes):

| Section | Merged before | After native-cache exclusion |
| --- | ---: | ---: |
| Type | 273 | 253 |
| Import | 659 | 659 |
| Function | 2,982 | 2,863 |
| Table | 5 | 5 |
| Memory | 4 | 4 |
| Global | 41 | 41 |
| Export | 33 | 33 |
| Element | 5,715 | 5,496 |
| Data count | 3 | 3 |
| Code | 2,955,663 | 2,881,606 |
| Data | 1,768,883 | 1,709,775 |
| Custom `go:buildid` | 114 | 114 |
| Custom `producers` | 71 | 71 |

The exact exported initial graph includes fingerprinted asset URLs. Compression
uses the same per-file gzip level 9 / Brotli quality 11 estimates as above:

| Artifact | Before raw / gzip / Brotli bytes | After raw / gzip / Brotli bytes |
| --- | ---: | ---: |
| Wasm | 4,734,489 / 1,391,823 / 1,054,156 | 4,600,966 / 1,356,326 / 1,027,188 |
| Initial runtime graph | 4,838,208 / 1,423,066 / 1,081,056 | 4,704,685 / 1,387,564 / 1,054,086 |

All existing optimized/portable raw, initial gzip (1.5 MB) and initial Brotli
(1.15 MB) budgets remain unchanged.

Both `make test-pages` and `WASM_OPT=off make test-pages` pass. The portable
wasm is 4,949,300 bytes; its initial graph is 5,053,019 raw / 1,396,602 gzip /
1,051,924 Brotli bytes.

Final validation completed at 04:10 EDT on 2026-10-08: both Pages builds,
`make test-web`, `node web/test/smoke.mjs --in-memory`, `go test -p 2 ./...`,
CLI E2E 138/138, `make test-nxn` and `make test-kociemba` all passed.
The CLI and all `./tools/...` executables were built into `dist/` before E2E.
