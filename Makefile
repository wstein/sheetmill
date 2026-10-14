# SPDX-License-Identifier: EUPL-1.2
.DEFAULT_GOAL := help

.PHONY: help fmt fmt-check test vet check race mod-verify test-assets test-hugo test-studio \
        check-hugo test-harness test-browser test-browser-corpus test-browser-studio fuzz diff-check verify-clean

help:
	@printf '%s\n' \
	  'make check                Formatting, Go tests, vet and whitespace (read-only)' \
	  'make fmt                  Format Go sources (writes files)' \
	  'make test / vet / race    Individual Go checks' \
	  'make mod-verify           Verify downloaded Go dependencies' \
	  'make test-assets          Verify brand assets with Python 3' \
	  'make test-hugo            Installed pinned Hugo integration tests' \
	  'make test-studio          Installed pinned Hugo lifecycle/race tests' \
	  'make test-harness         Browser harness unit tests' \
	  'make test-browser         Default reader browser journeys' \
	  'make test-browser-corpus  All three-sheet reader journeys' \
	  'make test-browser-studio  Authenticated Studio browser journeys' \
	  'make fuzz                 Three bounded parser/import fuzz checks' \
	  'make verify-clean        Fail if tracked files changed' \
	  '' \
	  'Hugo targets require SHEETMILL_HUGO=/absolute/path/to/pinned/hugo.' \
	  'Browser targets require pinned Node, npm dependencies and Chromium; see tests/e2e/README.md.' \
	  'No dependency installation, distribution builds or deployment targets.'

fmt:
	gofmt -w cmd internal assets

fmt-check:
	@files=$$(gofmt -l cmd internal assets) || exit $$?; \
	  if test -n "$$files"; then printf 'Run make fmt:\n%s\n' "$$files"; exit 1; fi

test:
	go test ./...

vet:
	go vet ./...

check: fmt-check test vet diff-check

race:
	go test -race ./...

mod-verify:
	go mod verify

test-assets:
	python3 -m unittest discover -s assets -p 'test_*.py'

# Go integration tests can skip without Hugo; fail here instead.
check-hugo:
	@case "$$SHEETMILL_HUGO" in /*) ;; *) printf '%s\n' 'Set SHEETMILL_HUGO to an absolute installed Hugo path.' >&2; exit 1;; esac; \
	  version=$$("$$SHEETMILL_HUGO" version) || exit $$?; \
	  pin=$$(cat .hugo-version) || exit $$?; \
	  case "$$version" in "hugo v$$pin "*|"hugo v$$pin+"*|"hugo v$$pin-"*) ;; \
	    *) printf 'Expected Hugo %s; got %s\n' "$$pin" "$$version" >&2; exit 1;; esac

test-hugo: check-hugo
	go test ./internal/render/hugo -count=1

test-studio: check-hugo
	go test -race ./internal/studio -count=1

test-harness:
	cd tests/e2e && npm run test:harness

test-browser: check-hugo
	cd tests/e2e && npm test

test-browser-corpus: check-hugo
	cd tests/e2e && SHEETMILL_E2E_CORPUS=all npm test

test-browser-studio: check-hugo
	cd tests/e2e && npm run test:studio

fuzz:
	go test ./internal/frontmatter -run '^$$' -fuzz FuzzParse -fuzztime 5s -parallel 2
	go test ./internal/markdown -run '^$$' -fuzz FuzzParse -fuzztime 5s -parallel 2
	go test ./internal/importer/reference -run '^$$' -fuzz FuzzConvert -fuzztime 5s -parallel 2

diff-check:
	git diff --check

verify-clean: diff-check
	git diff --exit-code
	git diff --cached --exit-code
