// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wstein/sheetmill/internal/importer/reference"
)

const importUsage = "usage: smill import reference <file> --repository <https-url> --source-path <repo-path> --revision <full-sha> --license <id> --attribution <text> (--out <new.md> | --dry-run) [--report json]"

func runImport(args []string, out, errout io.Writer) int {
	fail := func(message string) int { fmt.Fprintln(errout, message); return 2 }
	if len(args) < 2 || args[0] != "reference" {
		return fail(importUsage)
	}
	fs := flag.NewFlagSet("import reference", flag.ContinueOnError)
	fs.SetOutput(errout)
	var p reference.Provenance
	var destination, report string
	var dry bool
	fs.StringVar(&p.Repository, "repository", "", "source repository HTTPS URL")
	fs.StringVar(&p.Path, "source-path", "", "relative repository source path")
	fs.StringVar(&p.Revision, "revision", "", "full source commit ID")
	fs.StringVar(&p.License, "license", "", "source license identifier (does not approve redistribution)")
	fs.StringVar(&p.Attribution, "attribution", "", "source attribution")
	fs.StringVar(&destination, "out", "", "new canonical .md file; never overwritten")
	fs.StringVar(&report, "report", "", "json for machine-readable loss report")
	fs.BoolVar(&dry, "dry-run", false, "validate and report without writing files")
	if err := fs.Parse(args[2:]); err != nil {
		return fail(importUsage)
	}
	if fs.NArg() != 0 || (!dry && destination == "") || (report != "" && report != "json") {
		return fail(importUsage)
	}
	paths, err := inputs(args[1], false)
	if err != nil {
		return fail(fmt.Sprintf("%s:1:1: IO001: %v", args[1], err))
	}
	b, err := read(paths[0])
	if err != nil {
		return fail(fmt.Sprintf("%s:1:1: IO001: %v", args[1], err))
	}
	result, ds := reference.Convert(b, p)
	if len(ds) > 0 {
		for _, d := range ds {
			fmt.Fprintf(errout, "%s:%s\n", args[1], d.Error())
		}
		if report == "json" {
			b, e := result.Report.JSON()
			if e != nil {
				return fail("IO002: " + e.Error())
			}
			if _, e = out.Write(b); e != nil {
				return fail("IO002: " + e.Error())
			}
		}
		return 1
	}
	reportBytes, err := result.Report.JSON()
	if err != nil {
		return fail("IO002: " + err.Error())
	}
	if !dry {
		if err := createOutput(destination, result.Markdown); err != nil {
			return fail(fmt.Sprintf("%s:1:1: IO003: %v", destination, err))
		}
	}
	if report == "json" {
		if _, err := out.Write(reportBytes); err != nil {
			return fail("IO002: " + err.Error())
		}
	} else {
		for _, f := range result.Report.Findings {
			fmt.Fprintf(errout, "%s:%d:%d: %s: warning: %s\n", args[1], f.Span.Start.Line, f.Span.Start.Column, f.Code, f.Message)
		}
	}
	return 0
}

func createOutput(destination string, b []byte) error {
	if !strings.HasSuffix(destination, ".md") {
		return fmt.Errorf("output must be a new .md file")
	}
	abs, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	parent, info, err := checkedPath(filepath.Dir(abs))
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output parent must be an existing directory")
	}
	root, err := os.OpenRoot(parent)
	if err != nil {
		return err
	}
	defer root.Close()
	// O_EXCL rejects existing regular files, symlinks and directories atomically.
	f, err := root.OpenFile(filepath.Base(abs), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(b); err != nil {
		f.Close()
		return fmt.Errorf("output write failed; partial file may remain: %w", err)
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("output sync failed; file may remain: %w", err)
	}
	return f.Close()
}
