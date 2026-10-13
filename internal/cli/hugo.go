// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wstein/sheetmill/internal/markdown"
	"github.com/wstein/sheetmill/internal/render/hugo"
)

func buildCommand(out, errout io.Writer, status *int) *cobra.Command {
	build := &cobra.Command{Use: "build", Short: "Export a renderer project (no tool execution)"}
	var destination, title, base, editBase string
	command := &cobra.Command{Use: "hugo <file-or-directory>...", Short: "Export a self-contained Hugo project to a new directory", Args: cobra.MinimumNArgs(1), Run: func(cmd *cobra.Command, args []string) {
		*status = runHugo(args, destination, hugo.Options{Title: title, BasePath: base, EditBaseURL: editBase}, out, errout)
	}}
	command.Flags().StringVar(&destination, "out", "", "new project directory; never overwritten")
	command.Flags().StringVar(&title, "title", "SheetMill References", "reference library title")
	command.Flags().StringVar(&base, "base-path", "/", "URL path the site is served under, such as /sheetmill/")
	command.Flags().StringVar(&editBase, "edit-base-url", "", "GitHub edit URL base for the authored sheets (optional; never inferred from provenance)")
	command.MarkFlagRequired("out")
	build.AddCommand(command)
	build.RunE = func(cmd *cobra.Command, args []string) error { return fmt.Errorf("a renderer is required") }
	return build
}
func runHugo(args []string, destination string, options hugo.Options, out, errout io.Writer) int {
	seen := map[string]bool{}
	var files []string
	for _, arg := range args {
		paths, err := inputs(arg, true)
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", arg, err)
			return 2
		}
		for _, path := range paths {
			if !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
		}
	}
	sort.Strings(files)
	// Render identifies sheets by basename; one basename may name several paths.
	sourcePaths := map[string][]string{}
	var sheets []hugo.Sheet
	invalid := false
	for _, path := range files {
		b, err := read(path)
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", path, err)
			return 2
		}
		document, ds := markdown.Parse(b)
		for _, d := range ds {
			fmt.Fprintf(errout, "%s:%s\n", path, d.Error())
			invalid = true
		}
		name := filepath.Base(path)
		sourcePaths[name] = append(sourcePaths[name], path)
		sheets = append(sheets, hugo.Sheet{ID: sheetID(name), SourceName: name, Document: document})
	}
	if invalid {
		return 1
	}
	project, ds := hugo.Render(sheets, options)
	report := func(ds []hugo.Issue, suffix string) {
		printed := map[string]bool{}
		for _, d := range ds {
			paths := sourcePaths[d.Source]
			if len(paths) == 0 {
				paths = []string{"smill"} // project-level issue without a sheet origin
			}
			for _, path := range paths {
				line := fmt.Sprintf("%s:%s%s", path, d.Error(), suffix)
				if !printed[line] {
					printed[line] = true
					fmt.Fprintln(errout, line)
				}
			}
		}
	}
	report(ds, "")
	if len(ds) > 0 {
		return 1
	}
	report(project.Warnings, " (warning)")
	if err := writeFiles(destination, project.Files); err != nil {
		fmt.Fprintf(errout, "%s:1:1: IO003: %v\n", destination, err)
		return 2
	}
	if _, err := fmt.Fprintf(out, "Exported %d sheets to %s (Hugo %s; local generation is a separate command).\n", len(sheets), destination, hugo.Version); err != nil {
		fmt.Fprintln(errout, "IO002: "+err.Error())
		return 2
	}
	return 0
}
func sheetID(name string) string {
	name = strings.TrimSuffix(name, ".md")
	var b strings.Builder
	separator := false
	for _, c := range strings.ToLower(name) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(c)
			separator = false
		} else {
			separator = true
		}
	}
	if b.Len() == 0 {
		sum := sha256.Sum256([]byte(name))
		return fmt.Sprintf("sheet-%x", sum[:6])
	}
	return b.String()
}

// writeFiles lets tests inject a failure into the real writer.
var writeFiles = writeProject

func writeProject(destination string, files map[string][]byte) error {
	if destination == "" {
		return fmt.Errorf("output must be a new directory")
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
	name := filepath.Base(abs)
	// Mkdir atomically refuses all existing destinations, including symlinks.
	if err := root.Mkdir(name, 0700); err != nil {
		return err
	}
	project, err := root.OpenRoot(name)
	if err != nil {
		return err
	}
	defer project.Close()
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if !filepath.IsLocal(path) {
			return fmt.Errorf("invalid project path")
		}
		if err := project.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("partial project may remain: %w", err)
		}
		f, err := project.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return fmt.Errorf("partial project may remain: %w", err)
		}
		_, writeErr := f.Write(files[path])
		closeErr := f.Close()
		if writeErr != nil {
			return fmt.Errorf("partial project may remain: %w", writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("partial project may remain: %w", closeErr)
		}
	}
	return nil
}
