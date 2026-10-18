# SheetMill v0.2 plan: controlled import and offline text access

**Status:** revised proposal; documentation only, implementation requires separate
owner authorization. **Target:** `smill` v0.2.0, without a committed release date.
The broader cheatsheet hub remains a direction, not the v0.2 delivery contract.

## Scope and existing foundations

The proposed MVP adds controlled local TLDR import, a reusable plaintext renderer,
an offline local/stdin viewer, and explicit static text URLs. It builds on the
existing source format rather than introducing a new authoring language.

| Foundation | Current contract |
| --- | --- |
| GFM parsing | Goldmark v1.8.6 already provides GFM; see [source format](format/gfm-plus.md) |
| Shared command cards | [Spec 015](../specs/015-command-cards/spec.md), including shared parameters and legacy navi aliases, is implemented |
| Navi export | [Spec 006](../specs/006-navi-renderer/spec.md); `.cheat` generation, not navi ingestion or execution |
| Library navigation | [Spec 014](../specs/014-library-navigation/spec.md) is implemented |
| Sheet freshness | [Spec 017](../specs/017-freshness/spec.md) is implemented; it is not the navi profile |
| Controlled TLDR import | [Spec 013](../specs/013-tldr-import/spec.md) is draft, specification only; advance it before implementation |
| Release workflow | [Spec 019](../specs/019-release-workflow/spec.md) governs release procedure, not permission to release |

Proposed flow: local TLDR file plus explicit provenance → controlled converter →
canonical GFM+ → existing parser/validator → schema-1 IR → existing Hugo/navi
renderers or the proposed plaintext renderer. Local `smill view` consumes that
same parser and plaintext renderer. Static text is another presentation of the
same content; no source command is run at any stage.

## Compatibility and version boundaries

Keep `cheatsheet.cards.<id>.params` and the documented legacy declaration aliases.
Names remain `[a-z][a-z0-9_]{0,31}`; types remain exactly `string`, `int`, and `path`.
The shared declaration accepts required `type` and optional `description`, without
`bool`, defaults, choices or providers. Old/new declarations for the same card
coalesce only when their complete normalized meanings agree; conflicting values
are rejected, never partially merged. Preserve current card-ID rules, validation,
source positions, renderer behavior, complete successful `.cheat` bytes and pinned
entry identities; unknown metadata must still be preserved or diagnosed. The
[command-card contract](format/command-cards.md) remains authoritative.

CLI release version, source-format version, `cheatsheet.schema`, JSON IR `schema`,
and navi companion format version are separate contracts. This MVP proposes no
schema-2 change: source metadata schema and IR remain 1, and the companion version
stays unchanged. Existing renderers reject IR versions other than 1; do not
feed them a new version without an explicit compatibility change. Reserving new metadata requires a source compatibility decision
even if IR stays 1. A future IR v2 needs an explicit migration and renderer plan.

Existing freshness metadata remains optional. `cheatsheet.tool_version` is displayed
free text, not a parsed version range or runtime enforcement mechanism. Keep the
current [freshness rules](format/gfm-plus.md#sheet-freshness).

## Controlled TLDR import

Advance draft Spec 013 with these boundaries, preserving its provenance, loss and
refusal contracts. Proposed `smill import tldr <file>` accepts one explicit local
file, never URLs, globs, directories, recursive input or auto-detection. Require
`--repository`, `--source-path`, `--revision`, `--license`, and `--attribution` under
the existing [Reference import rules](imports.md). Require an exact revision for
corpus work; verify content and code licenses separately at that revision.
Conversion confers no redistribution permission and never fetches linked content.

Map the H1 to the title, introductory blockquotes to description and inert links,
and each example pair to one card under `## Examples`. Fix that section name for
the MVP; keep the original language in provenance. Use deterministic generated
card IDs and ordinary fenced command examples, without `.command` activation or
inferred shared parameters. Keep every command payload, including literal
`{{path/to/file}}`, `{{192.168.0.1}}`, `{{[-o|--output]}}`, escaped braces
and whitespace, byte-exact. These tokens need not be valid shared-parameter names.
Do not translate placeholders to `<name>`, tokenize shells, substitute values or
generate suggestions. Imported executable-looking metadata remains inert.

Derive platform and language only from validated `--source-path`, preserving the
upstream values under `source.original_metadata` as Spec 013 proposes. `common`
is a recorded upstream scope, not a host OS; `osx` stays `osx`, without silent
conversion to `darwin`. Recognized upstream scopes are preserved independently;
unrecognized layouts receive TLD102 and no guessed values. One file becomes one
sheet, with no common/platform merge or build-host selection. Long description
lines remain ordered sheet-description paragraphs; unsupported structures follow
the loss policy. Confirm these decisions in Spec 013 before coding.

Unknown, malformed or unsupported structures must be retained inert with
source-positioned loss findings, or rejected with no output when safe
representation is impossible. Prose brace syntax must not become heading
attributes. Validate converted bytes through the canonical parser before writing.
Keep Spec 013's proposed TLD diagnostics distinct from existing IMP diagnostics;
finalize and document codes before implementation. `--dry-run --report json`
writes no files, reports placeholders and losses, and uses deterministic reports.
Existing outputs are refused; no force mode. Preserve license and attribution in
the converted source and any reports that describe its origin.

## Plaintext renderer and offline viewer

Add one reusable IR-to-UTF-8 plaintext renderer, shared by the proposed
`smill view <local.md>`, `smill view -`, and static text generation. The viewer
reads local files or stdin offline: no remote lookup, link fetching, execution,
interactive substitution or automatic pager. Source limits and core validation
continue to apply. Validate the complete input before emitting rendered content;
send rendered text only to stdout, diagnostics to stderr and failures to a nonzero
exit status. Broken pipes must end promptly without a panic or diagnostic mixed
into stdout.

Specify deterministic ordering and layout before implementation. Use a fixed
80-column prose layout by default with an explicit bounded width option; do not
consult terminal size, locale or environment. Preserve command lines without
wrapping or truncation, apart from the safety escaping below. Define table/list
fallbacks in renderer goldens. TTY and piped output are identical for identical
options. MVP color is always off; emit no ANSI, OSC links or cursor controls.
Any later color mode needs a separate explicit contract.

Treat **every source-derived field** as untrusted display data: titles, headings,
prose, commands, code languages, links, metadata, provenance and diagnostics.
Escape C0/C1 controls (including ESC, BEL, CR and backspace), DEL and Unicode
format controls into visible notation; only renderer-controlled line breaks and
spacing may remain active. Apply the same safety boundary to stdout and stderr,
and test encoded/multibyte controls and unterminated escape sequences. Preserve
original source/IR bytes; a safely escaped display is not an executable or
byte-exact export of command payloads. Raw HTML remains inert visible text;
existing Hugo behavior reports REN101. HTML-comment directives are not already
recognized metadata and would require future parser and renderer work.

## Explicit static text URLs

Propose an opt-in `.txt` companion route for each sheet, with collision checks and
an explicit discoverable link from HTML. Fix its route and width as part of the
renderer specification before coding. Serve UTF-8 text with a documented
`text/plain; charset=utf-8` hosting configuration and no ANSI. Use the shared
renderer bytes so local and hosted output agree for the same layout options.
A user can explicitly request that URL with `curl`; static hosting performs no
User-Agent negotiation or request-time conversion.

Hugo's [output-format configuration](https://gohugo.io/configuration/output-formats/)
supports static plaintext output, but does not itself supply HTTP content
negotiation. The current embedded theme is under `internal/render/hugo/theme/`;
select and verify layouts against the repository's pinned Hugo version rather
than assuming `assets/hugo/layouts/` or a specific legacy template lookup path.
Implementation, local documentation/site generation, hosting activation and
publication each remain subject to the applicable owner authorization.

## Deferred decisions

| Follow-up | Decision required before implementation |
| --- | --- |
| Inline parameter comments | Separate proposal for grammar, duplicate precedence, attachment boundaries, source positions, import archival and compatibility; retain current frontmatter declarations meanwhile |
| Platform-aware authoring/filtering | Define deterministic explicit user selection and unscoped behavior; separate operating system from shell dialect and output profile; define inheritance/conflicts, unknown platforms and TLDR mapping before new classes are accepted |
| Static choices | Decide whether values are suggestions or an enforced enum, which validator/consumer enforces them, type/default rules and versioning; do not imply existing runtime enforcement |
| Dynamic providers | Require a consumer execution/security contract before enablement: dialect, output profile, context/cwd/environment, quoting/substitution, trust/consent, network/filesystem access, timeout/output limits and refusal when unsafe or unsupported |
| General adapter interface or IR v2 | Demonstrate a second concrete need and define loss, provenance and migration contracts first |
| TLDR export | Specify representable subset, platform/language selection, license handling and reported losses; no broad round-trip fidelity promise |
| cheat.sh import/export | Select a named, bounded upstream dialect, refusal cases and license: the [service README](https://github.com/chubin/cheat.sh/blob/master/README.md) describes aggregated sources, not one universal comment-plus-command grammar |
| Navi ingestion | Separate specification and authorization; no unscheduled import path in this MVP |
| Remote view, interactive runners, pager or color | Separate networking, trust and terminal behavior contracts; remain outside offline MVP |

Parser non-execution alone cannot make dynamic providers safe: the consumer runs
them later. The [navi syntax](https://github.com/denisidoro/navi/blob/master/docs/cheatsheet/syntax/README.md)
includes executable variable machinery, so inert payload forwarding is not a
substitute for the deferred execution contract. No provider wiring is scheduled.

## Milestones and acceptance

Sequence these milestones after scope and implementation authorization; dates
and effort estimates require an agreed design and measured baseline. For every
behavior change, first add a failing acceptance test, implement it, then refactor.

| Milestone | Deliverable and acceptance gate |
| --- | --- |
| M1: finalize contracts | Update draft Spec 013 and specify plaintext/CLI/static routes, safety and diagnostics. Review compatibility and deferred decisions before implementation. |
| M2: controlled importer | Independently authored synthetic fixtures and byte-exact Markdown/report goldens cover placeholders, CRLF, provenance, license retention, malformed constructs and inert attributes/HTML. Successful output reparses; every unsupported construct yields a loss or refusal. Dry-run, existing-output refusal and validation failure leave no files. Fuzz converter/parser boundaries for panics and unsafe successful output. |
| M3: shared renderer and local view | Goldens cover all node types, tables, Unicode and all source fields containing controls. Verify local/stdin parity, fixed/explicit widths, pipe behavior, stdout/stderr separation, no pager/network/execution and no emitted terminal controls. |
| M4: static companions | Verify text route collisions, HTML discovery, shared-renderer byte parity and unchanged existing Hugo/navi output and identities. Any site generation or deployment waits for explicit authorization. |
| M5: evidence and documentation | Record a pinned local corpus manifest, conversion dispositions and benchmark baselines; update format/import/CLI/Hugo/diagnostic docs and agent guidance to match implemented behavior. Review the exact candidate before any separately authorized release step. |

Corpus verification uses an explicitly selected local snapshot. Record the exact
upstream commit, content hashes, licenses, platform/language directories, included
and excluded paths with reasons, and the total input denominator. Classify every
input as faithfully converted, converted with reported loss, or rejected; publish
counts and representative findings for each class. A zero-exit conversion with
loss is not faithful success. Synthetic fixtures stay the default committed test
data; upstream fixtures require license review. The
[TLDR style guide](https://github.com/tldr-pages/tldr/blob/main/contributing-guides/style-guide.md)
is the format reference; pin its revision when settling Spec 013.

Measure import, parse and plaintext rendering separately with fixed fixtures,
byte sizes, toolchain, hardware and allocations. Establish repeatable baselines
before proposing regression thresholds; neither a 98% conversion claim nor a
1.5 ms per-sheet target has supporting measurements here. Documentation review
of this plan does not establish runtime behavior or release readiness.
