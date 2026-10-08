# Cube — a CLI Rubik's Cube Toolkit

A Go command-line toolkit for Rubik's cubes: a correct NxNxN move engine, full WCA
move notation, a CFEN state/pattern language with verification, move optimization,
and shortest-sequence pattern search.

> **Honest status:** the move engine, verification, optimization, and search all work
> and are covered by tests. **Kociemba solves a complete 3×3 by default.** The
> complete beginner solver and [lesson](./examples/beginner.md) remain available.
> CFOP solves a complete 3×3 with named stage playback. **Dimensions 2–7 now solve
> completely:** 4–7 use center/edge reduction with parity correction, then Kociemba.

## Status

**Works today**
- `cube learn` — a complete layer-by-layer 3×3 solve, with checkpoints, orientation
  guidance and interactive recovery (`next`, `moves`, `undo`, `reset`, saved CFEN)
- `cube solve` — fast complete 3×3 Kociemba solution; `--method beginner` retains
  the beginner solver; `--method cfop` solves Cross, F2L, OLL and PLL with case names.
  `--algorithm` remains a supported alias
- `cube solve --dimension N` — complete 2×2 through 7×7 solutions; 2×2 uses the
  3×3 corners path, 4–7 reduce centers and edges, including both even-cube parities
- `cube solve --optimal --time-limit 1s` — shortest face-turn solution if proved
  before the deadline; an explicit error if the search times out
- `--goal first-layer` retains the explicit white-layer sublesson
- Physical 3×3 state validation: color counts, centers, cubie identity, flips, twists, parity
- NxNxN move engine (2x2 through large N), all WCA notation, whole-cube rotations
- `cube twist` — apply moves, render the cube (ASCII / colored / Unicode)
- `cube verify` — check an algorithm against CFEN start/target states (wildcards supported)
- `cube optimize` — cancel/merge moves (`R R R` → `R'`)
- `cube find` — IDA* with cubie coordinates and admissible pruning tables for
  exact/wildcard 3×3 targets; sticker BFS for other dimensions or move alphabets
- CFEN search targets and concrete start states, restricted `--moves`, and
  checkerboard/cross patterns; returns one shortest sequence
- `cube lookup`, `cube show`, the CFEN utility commands, and the `verify-*` database tools

**Not implemented yet**
- CFOP color neutrality, extended cross and F2L look-ahead remain future work.
- Full solving above dimension 7 is unsupported. Big-cube solutions prioritize
  correctness and currently use hundreds or thousands of slice turns.
- The dump contains incorrect descriptions and missing cases; quarantined rows need
  curation, and some OLL/PLL cases use multiple verified database algorithms.
- Optimal search is exponential on deep states; wildcard heuristics are weaker
  than exact-state heuristics. Use Kociemba for general scramble solving.

Native measurements use **1,000 uniform legal states in one process**, seed
`2026100709`, one search thread, Go 1.26.2 and Mac background QoS
(`taskpolicy -b nice -n 15`). Every answer is replayed against the original
stickers and the input is checked for mutation. `make test-kociemba` gates
**100% at ≤20 turns, warm mean <50 ms and p99 <250 ms**, runs a first-ever plain
solve in a fresh process with an empty cache (**≤1 second and ≤10,000,000 cache
bytes**), and independently constructs/replays 200 Python physical states,
including all 24 grips. `make bench-kociemba` runs the warm sample alone;
`make test-kociemba-cold` runs the cold gate alone.

Default 3×3 solves keep searching until they find **≤20 face turns**, independent
of wall-clock time. The search never relaxes that length bound at a deadline.
If the fast reductions exhaust, a complete fallback enumerates phase-one depths
0–20 with unrestricted phase-two suffixes; any ≤20-turn solution splits at its
last non-subgroup turn. This guarantees finite coverage without imposing an
empirical node cap. Whole-cube grip rotations are counted separately.

Explicit CLI search flags select the timed API: `--time-limit` is a hard search
deadline, with table setup separate. An expired search returns an explicit error
unless it already has a solution within `max(20, --target-length)` face turns.
A target below 20 remains a stopping goal; expiry may return a ≤20 incumbent.
`KociembaSolver.Solve` provides the untimed default, while `SolveKociemba` honors
the explicit options. Correctness tests inject clocks rather than relying on
machine speed. The larger deterministic sample can be rerun with
`CUBE_DETERMINISM=1 go test -p 2 ./internal/cube -run '^TestTwoPhaseDeterminism10000$' -count=1 -v`.

A 10,000-state run after the deadline fix (same seed and policy) solved **100%
at ≤20 turns**, with **15.71 ms warm mean / 122.69 ms p99 / 1,658.64 ms max**.
The largest observed work was **17,504,500 DFS cursor operations**, at zero-based
state 5466. This is a sampled maximum, not a universal work or latency bound;
the default search has no node cap. Forced-clock regression case 198 used
**6,705,422 operations** and returned 20 turns.

The compact default measured **19.889 mean / 20 max turns, 100% at ≤20**, with
**12.29 ms warm mean / 106.08 ms p99 / 475.02 ms max**. The first-ever plain solve
of the cold-gate scramble took **214.45 ms**, including setup, and wrote
**3,337,787 bytes (3.338 MB)**. A fresh cached Go subprocess including a uniform
solve took **156.81 ms**. The independent 200-state CLI oracle measured **19.935
mean / 20 max turns**, **192.92 ms process mean / 821.07 ms max**, including
cached table setup in each process. These are seeded measurements under the
policy above, rather than worst-case bounds.

Default tables quotient phase-one twist/slice and slice/flip coordinates under
**16 U/D-axis symmetries (eight rotations and their mirrors)**, giving **168
twist classes and 45 slice classes**. Twist/flip uses the eight symmetries that
preserve flip independently of slice
(**324 twist classes**), with four-bit distances in two frames. An exact
**7,646-entry symmetry-reduced radius-five frontier** quickly rejects shallow
dead ends; tests check every descending witness and closure of all layers below
the frontier. Phase two has **2,768 corner and edge permutation classes**,
paired with slice permutation and the other group's four-piece combination
plus parity; inverse bounds preserve the existing correlation pruning.
Three axes, inverse views and up to three pre-moves share a resumable total-length
budget. A turn and its inverse share phase-one work; both phase-two endings are
checked. Reductions advance together by phase-one plus pre-move length.

The engine-owned **3,337,787-byte** compressed compact asset ships inside the
native and WASM executables. First use copies it to
`cube/coordinates-v5.bin.gz` in the user's cache; `CUBE_CACHE_DIR` overrides the
directory. A missing or corrupt cache uses the embedded asset without generating
large tables. `make generate-tables` rebuilds the asset from this engine's moves and exhaustively compares every
symmetry-reduced pair distance with unquotiented BFS. Checksums, dimensions and
move fingerprints guard the cache. The committed generated file is under 5 MB.

The cache files and their complete on-disk sizes are:

| File | Bytes | Used by |
|---|---:|---|
| `coordinates-v5.bin.gz` | 3,337,787 | Embedded compact tables for ordinary Kociemba and coordinate search |
| `coordinates-2x2-v1.bin` | 5,538,536 | Optional native 2×2 cache of the same tables with stored gzip blocks, generated on first 2×2 use |
| `edges-v1.bin` | 14,256,064 | Four-edge databases for CFOP and exact-state search, generated on first use |
| `phase1-sym8-v1.bin` | 147,502,144 | Optional large phase-one table, explicitly built with `cube tables build --large` |
| `optimal-v1.bin` | 134,568,064 | Full-corner and two six-edge databases, generated lazily for deep `--optimal` search |
| `optimal-phase1-sorted-sym16-cap11-v1.bin` | 1,021,870,144 | Saturated sorted phase-one database, generated lazily for deep `--optimal` search (974.53 MiB) |

All files live in the same cache directory. Ordinary Kociemba needs only the
compact file. Missing, unreadable, corrupt or non-regular cache files fall back
to the embedded compact asset or an in-memory rebuild of the requested database.
Optional writes use temporary files and atomic rename; an unwritable cache does
not prevent solving. The explicit `tables build --large` command requires
persistence and exits nonzero with a diagnostic if a write or rename fails.
Timed optimal setup checks its deadline throughout cache I/O and generation;
an expired budget returns an error without an unproved solution or partial cache.

The previous large-table solver measured **19.932 mean / 20 max turns, 100% at
≤20**, **5.68 ms warm mean / 57.22 ms p99 / 187.94 ms max**, with a **625.88 ms
fresh cached Go subprocess**. Its CLI oracle measured **19.935 mean / 20 max
turns**, **558.76 ms process mean / 890.13 ms max**. Its phase-one table had
**295,004,160 states / 147,502,080 packed bytes**; generation measured **120.99 s**
under background QoS with roughly **1.6 GiB** of temporary memory. A first-ever
plain solve regressed to about **37 s on performance cores / 161 MB of cache**.
That table is now optional: **`cube tables build --large`** explicitly persists
`phase1-sym8-v1.bin` (**147,502,144 bytes including its 64-byte header**),
and **`CUBE_LARGE_TABLES=1 cube solve ...`** explicitly selects it. Ordinary solves
ignore it even if it already exists. `make test-phase1-tables` checks the optional
table's admissibility, consistency and descending witnesses (maximum distance 12).
For comparison, c862379 with CPU profiling measured **19.801 mean / 21 max turns,
98.8% at ≤20**, **88.80 ms mean / 1000.08 ms p99 / 1000.52 ms max**.

The pruning/search strategy is informed by
[min2phase](https://github.com/cs0x7f/min2phase). A paired private reference run
used [min2phaseCXX](https://github.com/lilborgo/min2phaseCXX), built with installed
Clang (`-std=c++14 -O3`) inside ignored `.scratch/`, on the identical 1,000 fixtures
and process policy. It returned **19.746 mean / 20 max turns, 100% at ≤20**, with
**5.28 ms warm mean / 53.22 ms p99 / 316.64 ms max**, **991,712 cache bytes**,
**156.14 ms generation** and **24.73 ms fresh cached process**. Every answer
passed independent Python physical replay. Reference code and outputs are not
committed. `make export-reference-fixtures` writes the same 1,000 two-phase and
ten optimal URFDLB inputs into `.scratch/` without fetching code or dependencies.
Cube Explorer and nissy/vcube paired timings remain outstanding.

Optimal search retains the small databases for states within ten turns. Deep
`--optimal` states lazily build full-corner and two disjoint six-edge databases:
**88,179,840 corner states** and **42,577,920 states per edge group**, packed into
**44,089,920 + 21,288,960 + 21,288,960 bytes**. Factorized six-edge transitions
use **47,900,160 bytes**. Their checksummed `optimal-v1.bin` is
**134,568,064 bytes (128.33 MiB)**; earlier generation took **67.53 seconds**.

The stronger heuristic follows [Kociemba’s huge optimal coordinate](https://kociemba.org/math/optimal.htm):
it fixes the four slice edges in place as well as orienting all cubies. Sixteen U/D-axis symmetries reduce corner twists to **168 classes**,
with **4,087,480,320 combined entries**. The [modulo-three encoding](https://kociemba.org/math/pruning.htm) uses two bits
for exact distances through ten turns; remaining entries give the admissible lower bound
**eleven**. Saturation retains consistency across moves. The parent's bound
and a small lookup recover each child's bound, avoiding division in the search.
The cache is **1,021,870,144 bytes (974.53 MiB)**, including its checksum and move
fingerprint; the two optimal caches total **1,156,438,208 bytes**. Metadata and
moves are reconstructed from native coordinates and occupy about 1.2 MB.
Generation took **23m08.27s**, with **2,822,651,904 bytes peak RSS** and
**5,500,473,664 bytes peak memory footprint**, measured directly on the test
executable under Mac background QoS. A fresh cached setup took **1.75 s**;
a profiled warm search peaked at **1.20 GB RSS**.

Sorted-table allocation uses separate **1 MiB backing arrays**, with deadline
checks every **64 KiB** during filling, cache reads and hashing. This also bounds
the heap clearing on retries while preserving the existing cache bytes and
optimal bounds. With the smaller databases ready and the sorted cache missing,
a **200 ms** solver budget previously took **1.285 s**; it now returns the
explicit time-limit error in **200.04–200.12 ms** across three interrupted
attempts. A large cache load stopped in **200.03 ms**. Regression tests check
that cancellation publishes no partial table or cache, that a later cache
load still works, and that storage block boundaries preserve the cache format.
The four depth-16 fixtures retained identical search node counts.

Search takes the maximum of the corner, six-edge and three oriented sorted
phase-one bounds. Three equal positive phase-one bounds imply one additional
turn. Forward and inverse root bounds select the search direction. Within
eleven remaining turns, search also probes the inverse in three orientations.
Unknown entries give eleven; exact residues are rounded upward from the compact
phase-one lower bound. A tight inverse bound excludes next turns on its U/D
axis, because those turns would be redundant final subgroup moves in the
inverse solution. This follows the inverse move pruning described in
[Rokicki's nxopt design](https://github.com/rokicki/cube20src/blob/master/nxopt.md).
The strongest forward axis is checked first, and the sorted coordinate replaces
the unused slice-membership transition in the large-table path. At depth
thirteen and above, two workers partition the eighteen root moves; both must
finish to disprove a depth. Tables are shared and `GOMAXPROCS=2` bounds CPU work.
Compared with the previous phase-one heuristic, completed depth-fifteen searches
on the first two uniform fixtures visited **126,450 / 141,599 nodes**, versus
**957,974 / 1,077,891**: about **7.6 times fewer**. The two-worker measurements
were **0.306 / 0.432 s**, versus **2.78 / 3.78 s** for the smaller heuristic in
one worker. An eight-edge pattern-table experiment reduced nodes less and
increased elapsed time, so it is not part of the solver.

Completed depth-sixteen iterations on the first four uniform fixtures visited
**1,447,263 / 1,567,130 / 1,122,403 / 945,251 nodes**, versus
**1,695,840 / 1,895,546 / 1,397,075 / 1,157,518** before inverse pruning:
**14.7–19.7% fewer nodes**. The initial inverse-pruning trial took **11.25 s**
of search versus **18.17 s** for the baseline; these short timing comparisons
are sensitive to background scheduling and cache residency. Moving every
corner/six-edge lookup before the large-table lookups instead took **21.10 s**
and was dropped. The 147.5 MB exact phase-one fallback reduced visits by only
another **1.3–3.1%** without a consistent timing gain. Dynamic inversion within
the tree was slower (the first fixture rose from **4.07 s to 8.07 s**) and was
also dropped. Carrying final-move restrictions did not reduce visits.

A separate combined-coordinate experiment paired corner twist and U/D corner
membership with edge flip and slice membership under sixteen symmetries.
Its **9,930 corner classes / 10,066,636,800 entries** required
**2,516,659,264 cache bytes**. A five-byte chunked frontier stayed within the
memory budget, but generation hit its **40-minute limit** before finishing
depth ten. No cache was published, and the experimental coordinate is not in
the solver. The retained changes add no tables or cache bytes.

Large generation is opt-in through deep `--optimal` calls, with a one-time
stderr message. Ordinary solves retain the compact default tables and cache.
Optimal setup is included in `--time-limit`; interrupted builds never publish
partial tables. `make test-optimal` checks short distances against the separate
IDA* finder through ten turns, independently proves six fixed distances by
physical-sticker BFS, checks deeper finder fixtures, tests 14,000 known solving
suffixes against inverse pruning, and replays superflip in every grip.

The uniform-state benchmark (seed `2026100714`, **twenty states**, two search
workers, background QoS, no CPU profiling, **180-second limit per state**)
now proves **15/20**, versus **9/20** before inverse pruning. Five states remain
right-censored at 180 seconds. **Eight of twenty finish within 60 seconds**,
versus five previously. The **full-sample median is 91.02 seconds**, the mean of
the tenth and eleventh observations (87.02 and 95.02 seconds); both are observed,
so this median includes all twenty states without conditioning on success.
Eight proven answers have distance **17**, and seven have distance **18**.
Every answer passes full-cube replay and input-immutability checks. On the first
ten fixtures, **8/10** now finish, versus **5/10** previously.

| State | Proven FTM distance | Seconds | Large-search DFS nodes |
| --- | --- | ---: | ---: |
| 0 | 18 | 63.17 | 39,116,581 |
| 1 | 18 | 114.52 | 80,292,989 |
| 2 | 17 | 7.79 | 5,629,483 |
| 3 | 18 | 157.99 | 83,294,939 |
| 4 | 17 | 10.32 | 5,477,308 |
| 5 | 18 | 151.83 | 65,663,233 |
| 6 | 17 | 21.79 | 12,966,187 |
| 7 | 17 | 7.69 | 4,649,875 |
| 8 | unproved | >180 | 70,873,877 |
| 9 | unproved | >180 | 33,948,645 |
| 10 | unproved | >180 | 72,713,600 |
| 11 | 18 | 115.97 | 50,712,366 |
| 12 | 18 | 95.02 | 38,230,865 |
| 13 | 17 | 37.05 | 17,258,922 |
| 14 | 17 | 19.45 | 11,213,903 |
| 15 | unproved | >180 | 78,098,880 |
| 16 | 18 | 87.02 | 41,484,219 |
| 17 | 17 | 42.71 | 20,484,697 |
| 18 | unproved | >180 | 63,547,815 |
| 19 | 17 | 12.39 | 6,028,667 |

Node counts sum both workers and every IDA* iteration, including pruned DFS
entries; they exclude table setup and the preliminary short-state finder.
`make bench-optimal` logs each iteration's depth, visits and duration, including
the incomplete final iteration on timeouts. Cached setup took **9.67 seconds**
after the memory-intensive rejected table build; the timed benchmark command
reported **1,214,332,928 bytes (1.21 GB) maximum RSS**. Existing table generation
measurements remain **67.53 s** for corner/six-edge tables and **23m08.27s / 2.82
GB peak RSS** for sorted phase one; this run reused both caches.

**The practical optimal target is not met.** The median still exceeds sixty
seconds, and the three-minute censored trials do not establish completion within
ten minutes for all twenty states. `make bench-optimal` preserves the same
fixtures and budgets. A stronger combined coordinate that can be generated and
searched within the memory and time budgets remains useful future work. All
generated native caches and rejected experiments stay outside Git.

Superflip is recognized by exact cubie coordinates and uses
[Reid's published 20-turn lower bound](https://www.math.rwth-aachen.de/~Martin.Schoenert/Cube-Lovers/michael_reid__superflip_requires_20_face_turns.html)
plus a replayed 20-turn witness. General states exhaust every shorter IDA* depth
before an answer is returned; a timeout remains an explicit error.

NxN reduction measurements on the same Mac under background QoS (2026-10-08):
`make test-nxn` independently replayed **1,000 solutions**, with **100 uniform
legal states and 100 uniformly sampled single-layer scrambles per size**. States
are generated with independent integer 3D geometry; scrambles use 40×N turns.
The oracle covers all 24 grips, isolated and combined 4×4 OLL/PLL parity,
printed step counts and final CFEN. Every solve runs in a fresh CLI process.
Mean/max and times cover all 200 cases; the uniform column isolates the
100 uniform-state cases used for the move-length targets.

| Size | Mean / max moves | Uniform mean / max | Fresh CLI mean / max |
|---|---:|---:|---:|
| 2×2 | 18.79 / 20 | 18.66 / 20 | 56.63 / 119.07 ms |
| 4×4 | 92.72 / 128 | 92.16 / 115 | 138.46 / 286.04 ms |
| 5×5 | 192.65 / 253 | 193.19 / 253 | 178.48 / 691.74 ms |
| 6×6 | 457.31 / 522 | 452.05 / 508 | 172.16 / 462.87 ms |
| 7×7 | 638.53 / 712 | 640.83 / 712 | 162.76 / 369.73 ms |

Outer, numbered-slice, wide and half turns, and grip rotations each count
once. Times include process startup, coordinate cache loading and decoding embedded
reduction tables; the coordinate disk caches were already populated. Native 4–7
solves load the 3×3 cache alongside reduction. These are sample
measurements, not worst-case bounds. The oracle fails if either combined
or uniform mean moves exceeds the documented value by more than 5%, or
if this table disagrees with the solving guide. The 2×2 fresh-process mean must
also stay within **max(120 ms, 75% of the same-run compact 3×3 reference mean)**.
The oracle interleaves 20 fresh 3×3 solves of a short fixed scramble to allow
headroom for busy machines; isolated maxima do not gate this mean budget.
The former per-piece
reduction averaged 371.87, 532.82, 1058.29 and 1378.36 moves on sizes 4–7.

Profiling 100 cold 2×2 loads found 89% of sampled CPU in coordinate decoding,
including 68% in gzip inflation and 17% in gob decoding. The native 2×2 now
persists an optional 5.54 MB cache using stored gzip blocks, bound to the exact
embedded asset and checked by the same checksum, fingerprint and dimension
validation. This preserves every table and the search order; all 1,000 seeded
answers matched the previous executable move for move. Fresh 2×2 processes
measured **147.37 / 363.01 ms** before and **56.63 / 119.07 ms** after.
The first 2×2 use creates this cache; unavailable or corrupt caches fall back
to the compact asset. An empty-cache first solve took **247.44 ms**, with
**8,876,323 total cache bytes**. Interleaved old/new runs found no material
change on sizes 4–7 (6×6/7×7 differed by less than 2%). WASM retains its
existing loading path and embedded size.

Reduction uses center block searches on 4×4/5×5, batched bar commutators on
6×6/7×7, and slice-based edge pairing with short parity corrections.
Adjacent turns on one axis are canceled and packed into wide blocks.
Setup trees, optimized cycle costs, 5×5 center patterns and commutators use
1.5 MB of embedded compressed tables. Unit tests regenerate all four sizes and compare every byte
of their decoded content. Regenerate after changing generators with
`CUBE_GENERATE_NXN=1 go test -p 2 ./internal/cube -run '^TestNxNEmbeddedTables$'`.
All answers are verified, with no optimality guarantee. `--time-limit`
applies to the final Kociemba search.
See [the solving guide](./examples/solving.md#solve-other-sizes) for the reduction
method, parity examples and saved-state usage.

## Algorithm database and CFOP

The nine CSV files contain **159 rows** (including an empty `zz_misc.csv`). The
reproducible importer accepts **115**, merges **12** duplicate move sequences
across files/categories, and quarantines **44** with original rows and reasons
in [quarantine.json](./alg_dumps/quarantine.json). Together with the five existing
algorithms this produces **108 unique entries: 94 for 3×3, 14 for other sizes**.
Every entry has a concrete inverse-to-solved CFEN recognition pattern and passes
CFEN verification. The importer also checks that 3×3 OLL/PLL preserve F2L,
PLL preserves orientation, and F2L preserves the cross. Independent named
fixtures validate the 3×3 F2L/OLL/PLL case IDs. F2L descriptions and recognition
text derive corner/edge positions and sticker directions from those fixtures
in the standard FR frame before yaw/AUF. The importer does not independently
validate the claimed 2×2/big-cube case or parity description.

| Category | Entries |
|---|---:|
| F2L / OLL / PLL | 23 / 39 / 18 |
| Trigger / Advanced | 8 / 7 |
| Roux CMLL / LSE | 2 / 2 |
| 2×2 CLL / EG1 / EG2 / OLL / PBL | 6 / 1 / 1 / 1 / 2 |
| 4×4 / 5×5 / 6×6 parity | 2 / 1 / 1 |

Counts are category memberships; merged entries preserve aliases, categories,
and file/row provenance. **22 inverse pairs and 7 mirror pairs** are detected
from exact sticker permutations (including self-inverse/self-mirror cases).
Unknown relationships stay empty. Mirror means reflection in the left/right
plane; this does not claim every familiar case mirror matches without AUF.
Counts come from [import-report.json](./alg_dumps/import-report.json); a Go test
checks this section and TODO.md against the report so stale counts fail validation.

```sh
CUBE_CACHE_DIR=$PWD/.scratch/cube-cache taskpolicy -b nice -n 15 make import-algorithms
./dist/tools/verify-database
./dist/cube lookup --category F2L
./dist/cube solve "R U F2 L' B" --method cfop
```

CFOP uses `FindPattern` for a shortest white cross (at most eight face turns),
then recognizes safe one-slot F2L algorithms from the imported CFEN patterns,
with U setups and rotated slots.
When no database case matches, `FindPattern` inserts a pair while preserving the
cross and every solved slot. Last-layer recognition uses yellow-sticker masks
for OLL and full permutations for PLL. The valid database algorithms plus AUF
are composed into complete tables of **216 OLL orientations and 288 PLL
permutations**; incomplete raw case sets do not require a full-solver fallback.
Every checkpoint and the complete solution are replayed and verified.

The seeded 200-uniform-state Go oracle measured these real lengths on this Mac
under background QoS on 2026-10-07:

| Stage | Mean turns | Max turns |
|---|---:|---:|
| Cross | 5.790 | 8 |
| F2L 1 | 6.820 | 12 |
| F2L 2 | 6.715 | 12 |
| F2L 3 | 7.025 | 12 |
| F2L 4 | 7.350 | 12 |
| OLL | 10.635 | 18 |
| PLL | 13.070 | 20 |
| Total | **57.405** | **73** |

The same Go sample took **8.81 s total, 44.04 ms mean, 835.87 ms maximum**,
including first-use setup. A separate Python 3D geometry oracle replayed every
checkpoint on **200 independent uniform physical states**, including all 24
rigid grips: **58.275 mean, 75 maximum turns**. Fresh CLI processes averaged
**835.14 ms**, maximum **1796.34 ms** (167.03 s total, including table setup in
every process).
A face, wide or slice turn counts as one, including half turns; grip rotations
are excluded. Stage maxima can occur on different states. These samples are
not worst-case runtime or solution-length guarantees. This is paired F2L with
search fallback, without look-ahead or color neutrality.

## Quick Start

```bash
git clone https://github.com/ehrlich-b/cube
cd cube
make build            # builds dist/cube
make build-tools      # builds dist/tools/verify-algorithm and verify-database

# Apply moves and see the result
./dist/cube twist "R U R' U'" --color

# Finish a scrambled 3x3, one beginner checkpoint at a time
./dist/cube learn "R U F2 L' B" --interactive --color

# Return moves that solve all six faces
./dist/cube solve "R U F2 L' B" --headless
./dist/cube solve "R U F2 L' B" --target-length 19 --time-limit 500ms --headless
./dist/cube solve "R U F2 L' B" --method beginner --headless
./dist/cube solve "R U F2 L' B" --method cfop --headless

# Return moves that solve the first layer (other layers may remain scrambled)
./dist/cube solve "R U F2 L' B" --goal first-layer --headless

# Verify an algorithm against CFEN states (sexy move x6 = identity → solved)
./dist/cube verify "R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'" \
    --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"

# Optimize a sequence
./dist/cube optimize "R R R"            # → R'

# Search for a short sequence that reaches a pattern
./dist/cube find sequence "R U"         # → U' R'
./dist/cube find sequence "R U F2 L' B D2 R' F U2 L" --max-moves 10
./dist/cube find --target 'YB|Y9/?9/?9/?9/?9/?9' --start 'YB|Y9/R9/B9/W9/O9/G9'
./dist/cube find sequence R2 --moves R --max-moves 2  # → R R
```

See [examples/](./examples/) for checked CLI commands, the beginner lesson,
algorithm IDs and short verification/search workflows. Run `make test-docs`
to check the runnable CLI examples against the built binary.

## Interactive Website

The static [cube playground](./web/) uses the real Go engine compiled to
WebAssembly. There is no npm build step, server API, account, or external CDN.

```bash
make web
python3 -m http.server --bind 127.0.0.1 8080
# Open http://127.0.0.1:8080/web/
```

Turn the 3D cube with WCA keys (`R U F L D B`, `M E S`, `x y z`); hold Shift
for prime turns. Drag to orbit the view, or use arrow keys while the cube is
focused. The buttons, algorithm box, undo/redo, reset, scramble, and 2D net
work on phones too. Solving produces a verified sequence with play/pause,
step forward/back, clickable moves, a position slider, and adjustable speed.
The selector beside Solve defaults to Kociemba and also offers Beginner and
CFOP. CFOP playback groups moves into Cross, F2L 1–4, OLL and PLL with case names,
turn counts and skips; every move retains click, step and scrub playback.
The solution heading identifies the method that ran.
Kociemba uses the same 20-turn goal and one-second search budget as the CLI;
each worker initializes its own tables before starting that budget. Hard states
return the best solution found at the deadline. Browser smoke tests check that
Solve finishes and animation frames keep running during a cold-worker solve.
Lesson mode gives the beginner method's actual instructions and checks,
replanning the next hint from your current state after your own turns.

Search accepts a 3×3 CFEN target with `?` wildcards and a depth of 0–10. It runs
in a cancellable Web Worker with a 30-second limit; start with short searches.
Solver and lesson planning also run in workers. The WASM binding calls
`ParseMoves` / `ApplyMoves`, `GetSolver`, `PlanBeginner`, `FindPattern`, and
the existing CFEN and physical validation APIs. Search uses the CLI's shortest
IDA* path, including its sticker BFS fallback, without a separate JavaScript
solver or searcher.

Import/export uses concrete CFEN in YB storage order (`U/R/F/D/L/B`); actual
center colors preserve rotated grips. Share copies a URL hash containing the
scramble and algorithm, or an imported/manual starting state. Opening a link
restores the starting cube and prepares its algorithm for playback.

```bash
taskpolicy -b nice -n 15 make test-web
cd web && npm install && cd ..   # Playwright is only a web devDependency
taskpolicy -b nice -n 15 make test-web-smoke
```

The smoke test uses headless Playwright Chromium with a disposable profile,
starts Python on a random loopback port, stops it on exit, and saves
`.scratch/screens/{cube,search,cfop}-{1280x800,390x844}.png`. It uses the cached
Chromium on macOS; elsewhere install Chromium with `cd web && npx playwright
install chromium`. All test temporary files stay in `.scratch/`.

If an execution sandbox denies loopback sockets, `taskpolicy -b nice -n 15
node web/test/smoke.mjs --in-memory` runs the same browser assertions using
local-file request routing. This alternate transport does not check Python
serving; the normal smoke target continues to require the Python server. The
2026-10-07 validation used this transport because the sandbox refused loopback
bind; WASM API checks and the full desktop/mobile browser assertions passed,
including hard-state Solve completion and 55 animation frames during its cold
worker's setup/search. Go tests, vet, all 133 CLI E2E cases and both 200-state
independent Kociemba/CFOP oracles also passed under background QoS.

`make web-pages` rebuilds `dist/web` from a fixed list of runtime files. It
includes `index.html` and `.nojekyll`, fingerprints every other asset with a
SHA-256 digest of the complete release, and rewrites the HTML, module imports,
worker URLs, WASM URL and lazy solver data URLs to those relative filenames.
Changing JS, WASM or a solver asset gives the entire dependency graph new URLs,
including the Go runtime and integrity manifest.
Cached binaries from a previous release therefore cannot be paired with new
JavaScript. Pages controls caching of `index.html`; a cached document can still
show the previous release until it revalidates.

Run the publishing readiness gate without starting a server or publishing:

```bash
taskpolicy -b nice -n 15 make test-pages
```

It builds the exact Pages artifact, then serves only those static files through
headless Playwright request routing at `https://ehrlich-b.github.io/cube/`.
Unknown paths return 404 rather than falling back to the app. Direct links use
`/cube/index.html#...` or `/cube/#...`; the hash restores the cube and playback
without server-side routing. The harness checks relative paths and correct
`.wasm` / JS module MIME types, denies external requests and APIs, and supplies
no COOP/COEP headers. It verifies `crossOriginIsolated === false` and that
`SharedArrayBuffer` is unavailable while face turns, cold Solve and Search
workers still work. Workers receive a compiled WebAssembly module and create
their own runtime memory; they do not share a buffer.

Solver tables are separate assets. Kociemba and pattern search fetch the 3x3
coordinates on first use; reduction also fetches only the selected 4x4–7x7
table. Basic interaction, beginner lessons and CFOP need no table download.
Fetch integrity, SHA-256, sizes, coordinate dimensions and move fingerprints
protect the data. Failed/corrupt downloads preserve the cube and allow retry.
The browser coordinate format contains every native coordinate, with a checked
binary decoder instead of gob reflection. Native executables retain embedded
tables and their existing cache format. A second NxN wasm would duplicate the
Go runtime; the site shares compiled code and loads only data instead.

`make web` strips symbols and uses `-trimpath`. It also runs `wasm-opt -Oz` if
available on PATH or inside installed Emscripten. Set `WASM_OPT` to an explicit
executable, or `WASM_OPT=off` to use just Go; no build installs tools. Both paths
have size budgets in `test-pages`.

The gate also checks console/runtime errors, copied share links, control names,
DOM focus order, keyboard tab navigation, help, face/prime turns and playback,
and computed control text and focus-outline contrast. This is a basic control
accessibility pass, not a full screen-reader audit. A fresh 390×844 load uses
4× Chromium CPU throttling and must become interactive within 15 seconds.
Transfer is local and unthrottled, so this timing excludes an internet download.
The test reports raw, gzip (level 9) and Brotli (quality 11) byte estimates for
initial assets, each lazy asset, the whole export and WASM; actual Pages
compression is not assumed. It times first 3x3 and 7x7 solves, including lazy
loading and playback preparation, and tests missing/corrupt data recovery. It changes
WASM alone and JS alone and tests redeploys with poisoned old asset URLs.
Reports and screenshots stay in `.scratch/pages-readiness.json` and
`.scratch/pages-390x844.png`. The existing Playwright devDependency and a
provisioned Chromium are required; the test installs nothing. It can run in CI
with `make test-pages` (omit the macOS-only `taskpolicy` prefix on Linux).

The 2026-10-08 check (Go 1.26.2, Chromium 153, Mac background QoS) measured
**1.371 seconds to interactive** at 390×844 with 4× CPU throttling and local
transfer. Initial load requested **9,671,932 uncompressed bytes**; the worker
entry point loads on demand. The enabled controls passed the checks above,
with a minimum measured text contrast of **5.22:1** on the initial view.

| Pages export | Raw bytes | Gzip estimate | Brotli estimate |
|---|---:|---:|---:|
| All runtime assets | 9,673,579 | 5,009,496 | 4,530,967 |
| WASM alone | 9,573,518 | 4,979,772 | 4,505,423 |

To publish, run this **single command yourself** from the repository root,
after verifying that `origin` is the intended GitHub repository:

```bash
taskpolicy -b nice -n 15 make test-pages && taskpolicy -b nice -n 15 npx --yes gh-pages --dist dist/web --nojekyll
```

This builds and publishes only runtime assets to the `gh-pages` branch of the
current repository's origin after the readiness gate passes. GitHub Pages must
separately be configured to serve that branch's root, as described in
[GitHub's publishing-source documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/configuring-a-publishing-source-for-your-github-pages-site).
Publishing and Pages configuration are intentionally not performed by this
implementation task.

## Command Overview

| Command | Purpose | Status |
|---------|---------|--------|
| `twist` | Apply moves and render the cube | works |
| `verify` | Check an algorithm vs. CFEN start/target (`--start`/`--target`) | works |
| `show` | Render a cube with cross/OLL/PLL/F2L highlighting | works |
| `lookup` | Search the algorithm database | works |
| `optimize` | Cancel/merge a move sequence | works |
| `find` | Shortest exact/wildcard pattern search | works (3×3 IDA*, BFS fallback) |
| `parse-cfen` / `generate-cfen` / `verify-cfen` / `match-cfen` | CFEN utilities | works |
| `identify` / `show-alg` | Pattern identify / algorithm display | partial |
| `learn` | Teach a complete beginner solve, with recovery and checkpoints | works (3×3) |
| `solve --goal first-layer` | Solve the white first layer | works (3×3 beginner) |
| `solve` / `solve --goal full` | Kociemba; reduction for 4–7, corners for 2; beginner/CFOP on 3 | works (2×2–7×7) |
| `solve --optimal` | Prove a shortest face-turn solution within a time limit | works (deep states may time out) |

Note: `verify` takes a single positional argument — the algorithm — plus `--start`/`--target` flags.

## Move Notation

Full WCA (World Cube Association) notation:

| Type | Syntax | Description | Cube Sizes |
|------|--------|-------------|------------|
| Basic | `R`, `U'`, `F2` | Face moves (F/B/R/L/U/D) | Any |
| Slice | `M`, `E'`, `S2` | Middle-layer moves | Odd only (3x3, 5x5, …) |
| Wide | `Rw`, `Fw'`, `Uw2` | Multiple outer layers | 3x3+ |
| Layer | `2R`, `3L'`, `4U2` | Specific inner layers | 4x4+ |
| Rotation | `x`, `y'`, `z2` | Whole-cube rotations | Any |

Modifiers: `'` (counter-clockwise), `2` (double turn).

## Cube Orientation (canonical)

Yellow up, White down, Blue front, Green back, Orange left, Red right. The default CFEN
orientation is `YB` (yellow-up, blue-front). Apply `x`/`y`/`z` rotations before a sequence
to work from a different orientation.

## Testing & Invariants

```bash
make test        # Go unit tests, including the invariant suite
make e2e-test    # end-to-end CLI tests
make test-all    # both
make test-first-layer  # independent partial-goal regression oracle (Python 3)
make test-beginner     # independently replay full solutions and last-layer recovery
make test-kociemba     # 200 independent uniform physical states and move replay
make test-cfop         # 200 independent physical states, every CFOP checkpoint
make fmt && make vet   # before committing
```

The **invariant suite** is the project's safety net — it must stay green:

- `internal/cube/invariants_test.go` — sticker conservation, scramble+inverse = solved,
  determinism, and the **solver contract** (any non-empty solution must actually solve the cube;
  stubs SKIP rather than fake a pass).
- `internal/cfen/cfen_test.go` — canonical solved CFEN, cube↔CFEN round-trip (all orientations),
  wildcard matching, verify semantics.
- `internal/cli/commands_test.go` — no command registered twice.

## Architecture

```
cmd/cube/main.go                 # CLI entry point
internal/cli/                    # Cobra commands (twist, verify, solve, find, optimize, ...)
internal/cube/                   # Core engine
  cube.go                        # NxNxN representation, IsSolved, rendering
  moves.go / move_parser.go      # move parsing + application
  ring_generators.go / permutations.go  # the permutation engine
  algorithms.go                  # algorithm database
  solver.go / kociemba.go / cfop.go # full beginner, Kociemba and CFOP solvers
  nxn*.go                        # 2x2 corners, center blocks and wing pairing on 4x4-7x7
  coordinates.go / coordinate_tables.go  # cubie moves and deterministic pruning tables
  pattern_search.go / optimal_search.go  # shortest wildcard/exact search + BFS fallback
  first_layer.go / full_lesson.go / lesson_session.go  # beginner checkpoints and recovery
  state_validation.go            # physical 3x3 legality checks
  solving_db.go                  # experimental 4-look pattern matcher (currently unwired)
  cubie.go                       # piece addresses, ranges and 3×3 selector aliases
internal/cfen/                   # CFEN parsing, generation, conversion, matching
tools/                           # import-algorithms, verify-*, generate-patterns
```

## Programmatic Usage

```go
package main

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	c := cube.NewCube(3)
	moves, _ := cube.ParseMoves("R U R' U'")
	c.ApplyMoves(moves)

	fmt.Println(c.String())               // ASCII unfolded layout
	fmt.Println(c.StringWithColor(true))  // colored
	fmt.Println("solved:", c.IsSolved())

	cfenStr, _ := cfen.GenerateCFEN(c)    // YB|... state string
	fmt.Println(cfenStr)

	for _, alg := range cube.LookupAlgorithm("Sune") {
		fmt.Printf("%s (%s): %s\n", alg.Name, alg.CaseID, alg.Moves)
	}
}
```

## Roadmap

The complete beginner path now finishes all six faces. The unchanged full-solver
contract actively checks it. Partial first-layer results retain a separate result
type and explicit goal. Next improvements should follow a human physical-cube
trial. Database CFOP now complements Kociemba and bounded optimal search;
color neutrality, look-ahead, shorter big-cube solutions and dimensions above 7 remain future work.
See [TODO.md](./TODO.md) and [docs/solvers.md](./docs/solvers.md).

## License

MIT License — see [LICENSE](LICENSE).
