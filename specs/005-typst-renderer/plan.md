# Plan — 005

Specification only. Start implementation after separate authorization and after
002-gfm-parser is accepted. Build a renderer around canonical IR, maintaining portable
semantic metadata and renderer-owned layout. Pin the external tool before use.

First design a narrow, testable output contract; then write fixtures and tests;
implement in its own feature branch/PR. Actual Typst generation, distribution
builds and publication each remain separately gated. Do not create placeholder
CLI functionality before the renderer exists.
