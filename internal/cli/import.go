// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wstein/sheetmill/internal/importer/reference"
)

const importUsage = "usage: smill import reference <file> --repository <https-url> --source-path <repo-path> --revision <full-sha> --license <id> --attribution <text> (--out <new.md> | --dry-run) [--report json]"

type importOptions struct {
	provenance          reference.Provenance
	destination, report string
	dry                 bool
}

func referenceCommand(out, errout io.Writer, status *int) *cobra.Command {
	var options importOptions
	cmd := &cobra.Command{Use: "reference <file>", Short: "Convert a Reference sheet with explicit provenance", Run: func(cmd *cobra.Command, args []string) { *status = runImport(args[0], options, out, errout) }}
	// Parse errors keep the provenance flag guidance from the import usage.
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return usageError{err, importUsage}
		}
		return nil
	}
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error { return usageError{err, importUsage} })
	flags := cmd.Flags()
	flags.StringVar(&options.provenance.Repository, "repository", "", "source repository HTTPS URL")
	flags.StringVar(&options.provenance.Path, "source-path", "", "relative repository source path")
	flags.StringVar(&options.provenance.Revision, "revision", "", "full source commit ID")
	flags.StringVar(&options.provenance.License, "license", "", "source license identifier (does not approve redistribution)")
	flags.StringVar(&options.provenance.Attribution, "attribution", "", "source attribution")
	flags.StringVar(&options.destination, "out", "", "new canonical .md file; never overwritten")
	flags.StringVar(&options.report, "report", "", "json for machine-readable loss report")
	flags.BoolVar(&options.dry, "dry-run", false, "validate and report without writing files")
	return cmd
}

func runImport(input string, options importOptions, out, errout io.Writer) int {
	fail := func(message string) int { fmt.Fprintln(errout, message); return 2 }
	p, destination, report, dry := options.provenance, options.destination, options.report, options.dry
	if (!dry && destination == "") || (report != "" && report != "json") {
		return fail(importUsage)
	}

	paths, err := inputs(input, false)
	if err != nil {
		return fail(fmt.Sprintf("%s:1:1: IO001: %v", input, err))
	}
	b, err := read(paths[0])
	if err != nil {
		return fail(fmt.Sprintf("%s:1:1: IO001: %v", input, err))
	}
	result, ds := reference.Convert(b, p)
	if len(ds) > 0 {
		for _, d := range ds {
			fmt.Fprintf(errout, "%s:%s\n", input, d.Error())
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
			fmt.Fprintf(errout, "%s:%d:%d: %s: warning: %s\n", input, f.Span.Start.Line, f.Span.Start.Column, f.Code, f.Message)
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
