# SheetMill Studio: read-only local boundary

`smill serve` previews saved GFM+ through the existing Go compiler and installed
Hugo 0.167.0. It binds only assigned `127.0.0.1:<port>`; there is no public bind
flag. It provides no project writes, terminal, Git operation, source command
execution, arbitrary build arguments or remote import/fetch endpoint. Ordinary
source HTML is escaped, source shortcodes stay literal and external images remain
unfetched reference links. Navi/PDF/browser editing are absent.

## Authentication and requests

The CLI prints a cryptographic 32-byte one-use bootstrap secret. Successful
top-level GET consumes it, issues a random HttpOnly SameSite=Strict cookie and
redirects to the clean local root. Cookies have distinct port-specific names,
so concurrent previews do not displace each other's authorization. A restarted
session requires its new bootstrap. Cookies use local HTTP; this is not TLS.

Every generated asset, page, state and event request requires that session cookie,
exact assigned Host, valid Origin (when present), and compatible Fetch Metadata.
Cross/null origins, cross-site reads/embedding, alternate hosts and forwarding
headers are refused. Direct authenticated navigation without an Origin remains
usable. Bootstrap errors reveal no source or secret. Replayed/wrong tokens fail.
Only GET/HEAD are accepted; event streams require GET. No CORS permission is
added. Ambiguous encoded/traversing paths are refused before immutable-map lookup.

All responses carry no-store, no-referrer, nosniff, frame denial and a restrictive
local-only CSP: no inline scripts, forms, plugins or cross-origin framing. User-
activated ordinary external links are navigation, not automatic resource fetches.
The fixed browser opener receives only the CLI-generated local URL as arguments.

## Files, publishers and state

A retained `os.Root` grants only the selected file or direct directory. Ancestor
symlinks are rejected at startup; each read rechecks child type/inode/size/mtime
and compares two bounded reads. Atomic regular-file replacement is supported;
symlink/nonregular/unstable replacement is diagnosed. Bounds are explicit: 100
files, 8 MiB each, 32 MiB aggregate; excess selection is refused rather than
silently truncated. No source/Git file is written by Studio.

One publishing operation and one newest queued revision use owned temporary
roots. Renderer diagnostics remain per-source; valid peers update and invalid
sheets retain explicitly stale last-valid IR only when available. Latest observed
sequence wins; older completions cannot publish. Historical removals are bounded
and cannot mask renamed routes or resurrect deleted output after publisher failure.
If a retained snapshot's membership is stale, its aggregate library/search is
withheld while selected published pages remain available.

Hugo has fixed arguments, a fixed clock, controlled renderer configuration, isolated
HOME/cache/temp and no inherited HUGO/proxy environment. Its version check is
bounded to 5 seconds, generation to 30 seconds, combined output to 1 MiB and
publisher/snapshot byte budget to 64 MiB including retained prior output/IR.
Cancellation targets the owned child group with bounded escalation; no broad kill.
Session shutdown closes watcher, listener, root handles and SSE streams and removes
only owned temporary state. Go lifecycle tests verify the 5-second cleanup bound.

Private state contains relative source names, compiler spans/codes/messages,
monotonic state/source/publication identities and freshness, not credentials or
host paths. SSE allows 8 streams with 10-second heartbeats. HTTP-only trusted HTML
injection identifies the served publication, so the first state fetch and later
reloads cannot falsely mark older content current. Offline/error chrome shows
unknown freshness; reconnect fetches full state. A failed state fetch can recover
on the still-open stream without restarting the session.

## Verification and limits

Go tests cover authentication/replay/Origin/Host/forwarding/method/path boundaries,
concurrent cookies, streams, source resource ceilings, stable reads, route churn,
state ordering and owned-child cancellation. Actual-Hugo browser journeys cover
saved changes, invalid→repair, valid peers, atomic replacement/removal, single-file
scope, symlink denial, inert content, 100 rapid-save sequences, same-port restart,
offline/reconnect, first-state races and isolated state-request failure.

Authenticated Playwright tracing is disabled entirely. Bootstrap consumption is
outside browser history; startup stdout/tokens/cookies are not retained in test
reports. Failure screenshots and local sanitized logs stay in ignored `.local/`.
No artifact upload is configured. Source-relative diagnostics use DOM text APIs.

Hostile same-user software, a compromised OS/browser/toolchain and malicious local
administrators are outside this boundary. This is not a supported public server,
a filesystem sandbox for hostile same-user processes, or a measured performance/
five-author usability result. See [actual checks and gaps](studio-verification.md).
