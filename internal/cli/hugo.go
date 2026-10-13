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
	var destination, title, description, base, editBase, sourceBase, siteURL string
	command := &cobra.Command{Use: "hugo <file-or-directory>...", Short: "Export a self-contained Hugo project to a new directory", Args: cobra.MinimumNArgs(1), Run: func(cmd *cobra.Command, args []string) {
		*status = runHugo(args, destination, hugo.Options{Title: title, Description: description, BasePath: base, EditBaseURL: editBase, SourceBaseURL: sourceBase, SiteURL: siteURL}, out, errout)
	}}
	command.Flags().StringVar(&destination, "out", "", "new project directory; never overwritten")
	command.Flags().StringVar(&title, "title", "SheetMill References", "reference library title")
	command.Flags().StringVar(&description, "description", "", "home page introduction (optional)")
	command.Flags().StringVar(&base, "base-path", "/", "URL path the site is served under, such as /sheetmill/")
	command.Flags().StringVar(&editBase, "edit-base-url", "", "GitHub edit URL base for the authored sheets (optional; never inferred from provenance)")
	command.Flags().StringVar(&sourceBase, "source-base-url", "", "GitHub blob URL of the selection root; other relative links resolve to repository files")
	command.Flags().StringVar(&siteURL, "site-url", "", "canonical https://host/<base-path>/ of the published site; enables canonical links and a sitemap")
	command.MarkFlagRequired("out")
	build.AddCommand(command)
	build.RunE = func(cmd *cobra.Command, args []string) error { return fmt.Errorf("a renderer is required") }
	return build
}
func runHugo(args []string, destination string, options hugo.Options, out, errout io.Writer) int {
	project, count, report, status := renderSelection(args, options, errout)
	if status != 0 {
		return status
	}
	report(project.Warnings, " (warning)")
	if err := writeFiles(destination, project.Files); err != nil {
		fmt.Fprintf(errout, "%s:1:1: IO003: %v\n", destination, err)
		return 2
	}
	if _, err := fmt.Fprintf(out, "Exported %d sheets to %s (Hugo %s; local generation is a separate command).\n", count, destination, hugo.Version); err != nil {
		fmt.Fprintln(errout, "IO002: "+err.Error())
		return 2
	}
	return 0
}

// runStrict renders the selection in memory and reports every renderer warning
// (unresolved links, inert HTML, unfetched images) as an error. Nothing is written.
func runStrict(args []string, sourceBase string, errout io.Writer) int {
	project, _, report, status := renderSelection(args, hugo.Options{SourceBaseURL: sourceBase}, errout)
	if status != 0 {
		return status
	}
	report(project.Warnings, "")
	if len(project.Warnings) > 0 {
		return 1
	}
	return 0
}

// renderSelection reads, parses and renders the selected sheets. The returned
// report prints renderer issues against every selected path of their source.
func renderSelection(args []string, options hugo.Options, errout io.Writer) (hugo.Project, int, func([]hugo.Issue, string), int) {
	report := func([]hugo.Issue, string) {}
	seen := map[string]bool{}
	var files []string
	for _, arg := range args {
		paths, err := inputs(arg, true)
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", arg, err)
			return hugo.Project{}, 0, report, 2
		}
		for _, path := range paths {
			if !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
		}
	}
	sort.Strings(files)
	root := commonDir(files)
	// Render identifies sheets by basename; one basename may name several paths.
	sourcePaths := map[string][]string{}
	var sheets []hugo.Sheet
	invalid := false
	for _, path := range files {
		b, err := read(path)
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", path, err)
			return hugo.Project{}, 0, report, 2
		}
		document, ds := markdown.Parse(b)
		for _, d := range ds {
			fmt.Fprintf(errout, "%s:%s\n", path, d.Error())
			invalid = true
		}
		name := filepath.Base(path)
		sourcePaths[name] = append(sourcePaths[name], path)
		relative, _ := filepath.Rel(root, path)
		sheets = append(sheets, hugo.Sheet{ID: sheetID(name), SourceName: name, Path: filepath.ToSlash(relative), Document: document})
	}
	if invalid {
		return hugo.Project{}, 0, report, 1
	}
	project, ds := hugo.Render(sheets, options)
	report = func(ds []hugo.Issue, suffix string) {
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
		return hugo.Project{}, 0, report, 1
	}
	return project, len(sheets), report, 0
}

// commonDir is the deepest directory containing every selected file. Sheet paths
// are relative to it, so links between selected subdirectories resolve and no
// host path component reaches the output.
func commonDir(files []string) string {
	root := filepath.Dir(files[0])
	for _, file := range files[1:] {
		for rel, err := filepath.Rel(root, file); err != nil || !filepath.IsLocal(rel); rel, err = filepath.Rel(root, file) {
			parent := filepath.Dir(root)
			if parent == root {
				break // different volumes: no common directory
			}
			root = parent
		}
	}
	return root
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
