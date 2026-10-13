# Source and dependency provenance

- `LICENSE`: complete EUPL-1.2 text retrieved from the
  [SPDX license data](https://github.com/spdx/license-list-data/blob/main/text/EUPL-1.2.txt).
  [European Commission license portal](https://interoperable-europe.ec.europa.eu/collection/eupl/eupl-text-eupl-12).
  Original Go code carries SPDX EUPL-1.2 headers; original synthetic sheets/docs
  are part of SheetMill. No authored upstream cheatsheet is bundled.
- Goldmark v1.8.6: [upstream](https://github.com/yuin/goldmark), MIT; maintained GFM
  AST parser, chosen instead of a custom Markdown scanner. License stays in the module.
- go.yaml.in/yaml/v3 v3.0.5: [YAML organization fork](https://github.com/yaml/go-yaml),
  MIT/Apache-2.0 licensing per upstream; source nodes support strict duplicate and
  location validation. The archived gopkg.in module was not selected.
- Spec Kit CLI 1.1.3: already installed locally. `specify init` ran in scratch
  storage with `--integration codex --integration-options="--skills" --script sh
  --non-interactive --ignore-agent-tools`. Only generated `.specify` and `.agents`
  project infrastructure was copied after inspection. Its upstream MIT notice is
  retained at `.specify/LICENSE.spec-kit.txt`; this exception does not relicense
  Spec Kit scaffolding as EUPL. No package or global tool installation occurred.
  [Official initialization docs](https://github.github.io/spec-kit/installation.html).
- Go 1.27.2: supported release checked against [Go release history](https://go.dev/doc/devel/release).
  BundleDock `.go-version` pins 1.27.2 at `78da624484f57f84b097dd421ed61a8145427560`;
  WorkHarbor tests 1.27.2 at `5d6a8a423ec509383e1cf169bdea0a76d58ecd92`.
- Reference Python benchmark: inspected `source/_posts/python.md` at
  `82cbc0929f2d7a44e0bb5fc77e0a8442d0230b59`. The file stays in scratch storage,
  outside the public fixture tree. Parsing/conversion does not approve redistribution.
- GitHub Actions: checkout v5 `fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09` and
  setup-go v6 `924ae3a1cded613372ab5595356fb5720e22ba16` were resolved through
  official repository tag refs before pinning. No floating action tags are used.

No copied Reference implementation, CSS, authored sheet or sibling brand asset
is included. Import metadata records source licenses as supplied by users;
SheetMill does not infer legal permission from that identifier.

- Cobra v1.10.2: [official release](https://github.com/spf13/cobra/releases/tag/v1.10.2),
  Apache-2.0; explicitly selected by the user for CLI command/help/flag parsing.
  Transitive pflag v1.0.9 uses BSD-3-Clause and mousetrap v1.1.0 uses Apache-2.0;
  module checksums and original notices remain in their modules.
- Hugo 0.167.0: explicit local development tool, not invoked/downloaded by smill.
  The theme is original SheetMill source, with no upstream theme or Reference CSS.
  CI fetches the pinned official Hugo archive and verifies its SHA-256.
- Local web font and outlined brand fonts retain their separate SIL OFL notices;
  hashes and sources are recorded in [font provenance](../assets/fonts/README.md)
  and [asset reproduction](../assets/README.md).

- Browser E2E development package: `@playwright/test`, `playwright` and
  `playwright-core` 1.64.0, Apache-2.0, pinned with npm integrity hashes in
  `tests/e2e/package-lock.json`. No production Go/frontend dependency is added.
  Matching Chromium revision 1248 (156.0.8078.4) is provisioned explicitly from
  Microsoft's versioned Playwright tooling; browser binaries are not committed.
- Node 24.21.0 LTS is pinned for this private test package only. The official
  macOS arm64 archive was verified against the vendor's SHASUMS256.txt:
  `bed7eea5325e1108f32ce5228ddd6a5f0f08a499ee42aa7442aea583702f6057`.
  Local setup stayed in temporary storage. Node setup Action v6 was resolved through
  its official tag ref to `249970729cb0ef3589644e2896645e5dc5ba9c38` before pinning CI.
