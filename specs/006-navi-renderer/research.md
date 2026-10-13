# Source evidence — 006

Reviewed 2026-10-14. Documentation/source inspection only: no navi, shell,
finder, suggestion provider or example command was run. No runtime compatibility
result is claimed. The selected release is v2.24.0, commit
`5515367dc8a2a561d2d82b8352f729b1da4120d3`; use immutable links below rather than
assuming current master behaves identically.

| Upstream source | Observed behavior and effect on the proposal |
| --- | --- |
| [Syntax](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/docs/cheatsheet_syntax.md) | No provider is required for free input; `$` providers are commands. General GFM is not direct input. |
| [File discovery](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/filesystem.rs) | Accepts `.cheat` and `.cheat.md`; the Markdown extension does not change parsing rules. |
| [Parser](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/parser.rs) | `%`, `#`, `@`, `;` handling precedes fence handling. Fences do not protect arbitrary payloads. Entries are deduplicated by an item hash. |
| [Item identity](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/structures/item.rs) | Identity hashes concatenated trimmed tags and description without a separator; distinct contexts alone do not ensure distinct identities. Terminate contexts and check hashes before writing. |
| [Hash implementation](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/common/hash.rs) | Model the exact pinned item hashing, including its Rust string-hash input, rather than substituting an arbitrary hash. Reject duplicate hashes across an export batch. |
| [Deserializer](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/deser/mod.rs) | Placeholder recognition is broader than the proposed lower-case names; newline representation uses a control character. Do not infer shell syntax from angle-bracket balancing. |
| [Terminal serialization](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/deser/terminal.rs) | Finder fields use a separator containing U+2800; reject it in command/display text. |
| [Actor](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/commands/core/actor.rs) | Values, including environment values, are substituted as raw text. Providers spawn a shell. A first word `navi` is rewritten to an absolute executable path. Print mode is not a general no-execution boundary. |
| [Shell](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/common/shell.rs) and [configuration](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/config/mod.rs) | Execution uses the configured shell; fence language does not configure it. |
| [Variable map](https://github.com/denisidoro/navi/blob/5515367dc8a2a561d2d82b8352f729b1da4120d3/src/structures/cheat.rs) | Providers are keyed by context/name and replace previous definitions. Providers and shared contexts are deferred. |

## Decisions and pending evidence

- Select `bash-free-input-v1`: one physical Bash command per explicitly marked,
  identified card, card-local free-input hints and no executable providers.
- Treat parameter values as trusted raw shell text, not safely quoted arbitrary
  data. Record this limitation in generated headers and future user documentation.
- Keep current generic metadata preservation and `.navi` behavior unchanged.
  A future validator/schema and command projection are implementation tasks.
- Require pure parser/interpolation fixtures, independently authored examples,
  Bash grammar checks and a transitive no-execution audit before calling the
  profile tested. A model written from this evidence must document which pinned
  upstream branches it covers; do not describe it as running upstream navi.
- The IR does not currently expose a dedicated explicit-ID flag. Before implementing
  selection validation, determine a reviewed additive IR representation for this
  source fact; do not reparse Markdown in the exporter or infer explicitness from
  a generated slug. This prerequisite is included in T004.
- Multiline scripts, escaping literal placeholders, other shells and provider
  capabilities require subsequent scope/design approval.
