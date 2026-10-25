# Feature Specification: navi variable sources

**Feature Branch**: `codex/spec020-navi-variable-sources` (specification); implementation on a `codex/` branch
**Created**: 2026-10-20
**Status**: Implemented on main (specification #60, core #63, export #64, all merged 2026-10-20); issues #61 and #62 remain open
**Input**: Owner decisions of 2026-10-20: the Markdown sheet stays the single source;
`smill build navi` exports per-parameter `choices` and `source` as navi `$` lines.

## Scope and boundary

- The Markdown sheet is the single source. `.cheat` files remain generated output
  (export only; no `.cheat` ingestion in this feature).
- Parameters gain two optional, mutually exclusive fields, `choices` (a fixed list of
  literals) and `source` (a shell command line), plus an optional `options` mapping
  with navi argument options. They live in `cheatsheet.cards.<id>.params.<name>`
  (legacy alias `cheatsheet.navi.cards.<id>.params.<name>`).
- Sheet sources are treated like the contents of a navi cheat repository. There is no
  opt-in flag, no `--allow-dynamic` and no gating. navi runs a `$` line when its
  placeholder is filled.
- Boundary: SheetMill writes text only. It never runs navi, a shell or any command,
  and a lint test keeps the renderer free of `os/exec`, `syscall`, `plugin` and
  network imports (unchanged from 006).
- A `source` is exported unchanged. navi exports each chosen variable value as an
  environment variable, so an explicit `$name` reference in a `source` is passed
  through as shell text and is not analyzed (dependent variables are out of scope).
- Out of scope: dependent variables (`$ b: cmd <a>`, `$a` references), `@`
  shared contexts, `--fzf-overrides`, `--query`, `--filter`, `--header`, `--preview`,
  `--preview-window`, `--expand`, `--map` (it runs navi's own shell helper functions
  in non-fish shells, is experimental upstream and is unusable under nu), `.cheat` ingestion, other renderers consuming
  choices. These are possible later features.

## User Scenarios & Testing

### User Story 1 - Offer a fixed list of values (Priority: P1)

An author declares `choices: [main, develop]` for a parameter and gets a navi
suggestion list, with no shell knowledge needed to quote the values.

**Why this priority**: The most common variable need, and the one SheetMill can
generate safely from literals.
**Independent Test**: Export one card per profile and compare the complete `.cheat`
bytes with the golden files; feed the output to the pure Go model of navi's line grammar.

**Acceptance Scenarios**:
1. **Given** `choices: [main, "it's", süd]`, **When** exporting with each `--shell`
   profile, **Then** the card is followed by one `$ name: ...` line whose quoting
   follows that profile and which lists the values in authored order.
2. **Given** a choice containing a quote, backslash, `$`, `%`, a leading dash or
   non-ASCII text, **When** exporting, **Then** the value reaches the line quoted for
   the profile and is not altered.
3. **Given** a choice with a newline or control character, **When** checking, **Then**
   a positioned error rejects the sheet before any file is written.

### User Story 2 - Offer values from a command (Priority: P1)

An author declares `source: git branch --format='%(refname:short)'` and optional
`options` such as `column` or `multi`. The export writes the command and options
unchanged on one `$` line.

**Why this priority**: Dynamic suggestions are what navi users expect from `$` lines.
**Independent Test**: Golden exports with and without `options`; negative table for
text that cannot form a valid line.

**Acceptance Scenarios**:
1. **Given** a `source` and `options: {column: 2, header_lines: 1}`, **When**
   exporting, **Then** the line is `$ name: <source> --- --column 2 --header-lines 1`
   in the fixed option order of the data model.
2. **Given** a `source` containing `---`, a navi `<name>` token, a trailing backslash
   or an unterminated quote, **When** exporting, **Then** the card is skipped with a
   warning that names the parameter (`--strict` fails the build).
3. **Given** both `source` and `choices` on one parameter, **When** checking,
   **Then** a positioned error rejects the sheet.

### User Story 3 - Existing sheets are unaffected (Priority: P1)

A sheet that uses neither field exports byte-identical `.cheat` and `.navi.json`
files and the same pinned entry hashes as before.

**Why this priority**: Identity and output stability were a hard requirement of 015.
**Independent Test**: Existing goldens, migration fixtures and hash vectors pass
unchanged; sheets with sources keep the same entry hashes as without.

**Acceptance Scenarios**:
1. **Given** any sheet valid before this feature, **When** exporting, **Then** the
   bytes equal the pre-feature bytes.
2. **Given** a card with a source, **When** comparing its pinned hash with the same
   card without it, **Then** the hashes are equal and NAV008 behaves as before.

### Edge Cases

- `choices: []` is an error. Duplicate choices are an error. A choice that is empty,
  whitespace-padded or not a string (numbers, booleans, null) is an error: no coercion.
- `type: int` requires every choice to be a decimal integer string (`"3"`, `"-1"`).
- Choices with quotes, backslashes, `$`, `%`, backticks or non-ASCII letters are valid
  and quoted per profile. A choice containing `---` or a navi `<name>` token cannot be
  carried by navi's line grammar: the card is skipped with a warning.
- Fish quotes inside single quotes with `\'` and `\\`; bash and zsh close the quote
  (`'\''`); nu uses double-quoted strings. See the contract.
- `source` is one line without controls or surrounding whitespace. Unbalanced quotes
  use the profile's existing quote analysis (Bash/Zsh parser, Fish heuristic, none
  for nu) and skip the card.
- A parameter with a `source` or `choices` that its command never uses stays the
  existing NAV104 warning (card skipped). A used placeholder without declaration stays
  NAV104.
- A placeholder repeated in the command yields one `$` line.
- `options` without `source` or `choices` is an error; `column`, `delimiter` and
  `header_lines` require `source`; `multi: true` with `prevent_extra: true` is an error.
- A card with no `source`/`choices` exports exactly as before (free input).
- Imported Reference sheets: upstream `cheatsheet.cards` stays inert provenance (015).

## Requirements

### Functional Requirements

- **FR-001**: Shared parameter declarations MUST accept optional `choices`
  (nonempty list, 1 to 64 distinct strings of 1 to 128 code points), `source`
  (nonempty single-line string) and `options` (closed mapping), in addition to
  `type` and `description`. Unknown keys MUST remain errors (CMD001).
- **FR-002**: `source` and `choices` on one parameter MUST be an error (CMD001).
  The legacy alias accepts the same fields with the same rules and reports NAV006.
- **FR-003**: Choices MUST be strings, never coerced. They MUST NOT contain controls
  (including newline and tab), Unicode format characters, U+2028/U+2029, U+2800 or
  surrounding whitespace. For `type: int` each choice MUST match `^-?[0-9]+$`.
- **FR-004**: `source` MUST NOT contain controls, format characters, U+2028/U+2029,
  U+2800 or surrounding whitespace. Shared validation checks shape and these
  characters only; it does not analyze shell syntax.
- **FR-005**: `options` allows only `column` (integer 1 to 255), `delimiter` (nonempty
  string), `header_lines` (integer 0 to 255), `multi` and
  `prevent_extra` (booleans). Strings follow the character rules of FR-004.
  `options` requires a `source` or `choices`; `column`, `delimiter` and `header_lines`
  require `source`; `multi` and `prevent_extra` MUST NOT both be true.
- **FR-006**: The exporter MUST write, for each used parameter that has `choices` or
  `source`, exactly one line `$ <name>: <command>[ --- <options>]` directly after the
  card's command lines, in order of the parameter's first placeholder in the command.
  The card's text MUST contain no blank line and no newline inside a `$` entry.
- **FR-007**: For `choices` the exporter MUST generate the command from the literals
  with the quoting of the selected profile (contract), preserving order and values
  exactly. `source` MUST be exported byte for byte.
- **FR-008**: A card MUST be skipped with a warning, never rewritten, when a source
  or choice cannot be carried by navi's line grammar: it contains `---`; it contains a
  token matched by navi's placeholder regex (dependent variables are out of scope);
  a `source` ends with a backslash or has an unterminated quote under the profile
  check; an option string contains `---` (navi splits the whole remainder of the
  line on it) or, for simplicity, `'`. `NAV108` covers `source` and `options`,
  `NAV109` covers `choices`.
- **FR-009**: Every `$` line MUST name a declared parameter used in the command, and
  every declared parameter MUST be used (existing `NAV104`). No `$` line may start a
  second entry or alter the card's command text.
- **FR-010**: Pinned entry hashes, contexts, tags, titles, entry order, header and
  format version MUST NOT change. Cards without these fields MUST export byte-identically.
  NAV008 collision detection is unchanged because `$` lines are not part of the
  hash input.
- **FR-011**: The schema-1 IR MUST NOT change. Parameter hints stay in the generic
  metadata map; the `.navi.json` companion preserves them verbatim.
- **FR-012**: `schema/cheatsheet.schema.json` MUST describe the new fields and a test
  MUST keep it consistent with the validation rules.
- **FR-013**: No navi, shell, finder or source command may run in tests or the
  renderer. Tests compare bytes and use a pure Go model of navi's variable-line
  grammar.
- **FR-014**: The Hugo renderer and other targets MUST ignore the new fields; their
  output MUST NOT change.

### Key Entities

- **Variable source**: a parameter's `choices` or `source`, exported as one `$` line.
- **Argument options**: the closed `options` mapping, exported after `---`.
- **Profile quoting**: the per-shell rule that turns literals into a command.

## Success Criteria

### Measurable Outcomes

- **SC-001**: For each profile the committed golden `.cheat` file equals the export
  byte for byte, and the grammar model recovers exactly the declared variables,
  commands and options.
- **SC-002**: Every pre-existing navi golden, migration fixture and hash vector passes
  unchanged; no `.cheat` or `.navi.json` byte changes for sheets without the new fields.
- **SC-003**: Every row of the negative tables yields its documented code and
  position, and an error writes no file.
- **SC-004**: A property test over generated choice lists (including quotes, newlines
  rejected, unicode) finds one `$` line per variable in the model and the choice list
  unchanged after parsing the quoting back.
- **SC-005**: Mutation probes (tasks T014) are each detected by at least one test.

## Assumptions

Behavior is based on reading navi's source and docs at the version pinned in
[research.md](../006-navi-renderer/research.md) (the exact pin is read from the
repository's existing evidence, not repeated in tests). Anything about runtime
behavior that source reading cannot establish is listed as to-be-measured in
[research.md](research.md); no real navi run is claimed. Implementation starts
after the owner has merged this specification (grant of 2026-10-20 in AGENTS.md).
