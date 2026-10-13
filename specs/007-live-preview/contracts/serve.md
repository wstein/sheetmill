# Later-stage local serve contract

Not an MVP prerequisite. Maps to US3–US5; no implementation is authorized by this
design document. The one-sheet static publishing journey must pass first.

## CLI

Proposed form: `smill serve [file-or-directory] [--open] [--port <0..65535>]
[--hugo <absolute-executable>]`.

- Cobra parses at most one selection. Default selection is current directory,
  nonrecursive `.md` entries, using existing regular-file/symlink boundaries.
- Default port 0 chooses an available IPv4 loopback port. No bind-host/public flag.
- Resolve Hugo explicitly from `--hugo` or the installed `hugo` on PATH; require
  version 0.167.0 before preview work. No installation/download or source-defined tool.
- Print one usable local bootstrap URL and selected scope. `--open` launches that
  URL only after listener readiness; launch errors print the address and keep serving.
  Browser launch is a fixed platform operation with argument arrays, never source text.
- Invalid source at startup is a running diagnostic session, not fatal startup.
  Bad selection, missing tool, occupied explicit port or invalid configuration exits 2.
  Source invalidity during a running session is visible state; normal interruption
  exits 0 after cleanup, fatal runtime infrastructure failure exits 2.

## Read authorization and HTTP boundary

Use standard-library HTTP with exact Host validation against the assigned
`127.0.0.1:<port>`, rejecting alternate hosts, forwarded-host tricks and rebinding.
Validate Origin and Fetch Metadata on every private read/event request, including
GET. CrossOriginProtection may supplement this but is not sufficient for read
privacy. Reject unrelated/null origins and cross-site embedding; no permissive CORS.
Direct top-level navigation without an Origin is allowed only under session auth
(or for the non-sensitive bootstrap response below).

Proposed bootstrap: a cryptographically random 32-byte one-use secret in a CLI-
issued bootstrap URL. Successful top-level bootstrap consumes it, sets a random
HttpOnly SameSite=Strict session cookie, and redirects to the clean local root.
Bootstrap responses carry `Referrer-Policy: no-referrer` and `Cache-Control: no-store`.
Never include private source or credentials in unauthenticated bootstrap errors.
Session cookie auth and exact origin/host checks protect all preview/assets/state/
events. Cookie is local HTTP and not a claim of protection against a malicious
same-user process; hostile local software and OS compromise are outside this boundary.
Credential issuance changes only session state, never project files; no project
mutation route exists. Do not log bootstrap paths, cookie values or authorization
headers; authenticated Studio browser contexts disable tracing entirely because request
traces can expose cookies after bootstrap too. Failure screenshots and sanitized
logs provide browser evidence; screenshots must not show a bootstrap address.

Only GET/HEAD are served (event stream GET); other methods return 405. Generated
site routes come from an immutable file map, never a path joined to user files.
No raw repository browser, terminal, Git, arbitrary build command, remote URL fetch
or user-provided Hugo configuration. Reject traversal, invalid encodings,
backslashes, unknown routes and symbolic-link changes. Source names in diagnostics
are selection-relative; host paths remain terminal-only.

Trusted Studio client/status assets are under `/studio/`; private JSON state is
`GET /studio/state`, events `GET /studio/events`. Generated site pages/assets retain
their root routes. In HTTP HTML responses only, add the trusted local Studio client
script; generated files remain unchanged. Client adds fixed status chrome with
`textContent`/DOM APIs, never inserts diagnostic/source strings as HTML. Content
namespace remains `ref-`, existing app IDs remain `app-`, Studio IDs use `studio-`.
Use `nosniff`, no-referrer, no-store private responses and a restrictive local-only
CSP with no inline scripts, plugins, forms or cross-origin framing. Ordinary
user-activated external links remain navigation, not automatic resource fetching.

## State and events

State includes session sequence, selected sheet IDs/titles, per-sheet status,
source-relative diagnostics with spans, published revision and connectivity. It
contains no session credentials, raw unselected files or canonical schema additions.

SSE sends revision notifications and heartbeats every 10 seconds. A notification
carries only monotonically increasing sequence/state identity; client fetches
current state. Reconnection fetches a fresh full state; event replay is not required.
A disconnect marks output visibly stale immediately. Latest revision wins even
when earlier processing finishes later. Browser can reload its selected generated
page after publication, preserving the URL fragment if that card still exists;
missing/deleted references show an explicit unavailable state and navigation.

Polling: content hashes every 100 ms, 100 ms quiet coalescing, max 500 ms postponement.
Recheck regular-file/containment constraints each read. A stable observed revision
uses bounded source snapshots; atomic replace is supported, symlink replacement
is refused. Each new revision cancels obsolete publishing or supersedes its result.
Only one publisher and one latest queued revision; source-defined commands never run.
Hugo uses controlled renderer output and fixed arguments/config in owned temporary
roots. Tool execution and failure are observable; no silent fallback to a JS renderer.

Invalid sheets retain explicitly stale last-valid IR/output only when available;
valid peers still update. Duplicate routes mark all conflicting sources invalid.
Deleted sheets are unavailable. Limits and lifecycle follow [data-model.md](../data-model.md):
explicit failures, no silently omitted inputs, 5-second graceful shutdown and
escalation limited to the known owned Hugo process.

## Validation

Before enabling this stage, failing Go/HTTP tests cover scope, origin/host/token
checks, all request methods, resource ceilings, cancellation/order and cleanup.
Extend the same browser harness with saved changes, invalid→repaired source,
rapid saves, disconnect/reconnect and atomic/symlink replacement. Measure live
latency separately from compiling the temporary CLI and installing prerequisites.
Do not add PDF/navi tabs, browser editing, autosave, playgrounds or deployment.
