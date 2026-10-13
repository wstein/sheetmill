// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wstein/sheetmill/internal/render/navi"
)

func naviCommand(out, errout io.Writer, status *int) *cobra.Command {
	var destination, shell, root string
	var strict bool
	command := &cobra.Command{Use: "navi <file-or-directory>...", Short: "Export marked cards as navi .cheat files to a new directory", Args: cobra.MinimumNArgs(1), Run: func(cmd *cobra.Command, args []string) {
		*status = runNavi(args, destination, shell, root, strict, out, errout)
	}}
	command.Flags().StringVar(&destination, "out", "", "new directory for the .cheat files; never overwritten")
	command.Flags().StringVar(&shell, "shell", "bash", "shell profile navi is configured with: "+strings.Join(navi.ProfileNames(), ", ")+" (zsh, fish and nu are experimental: unverified against real shells)")
	command.Flags().StringVar(&root, "root", "", "directory source paths are made relative to (default: the common directory of the inputs); relative paths name each card to navi")
	command.Flags().BoolVar(&strict, "strict", false, "fail without writing when any card is skipped or any warning is reported")
	command.MarkFlagRequired("out")
	return command
}

func runNavi(args []string, destination, shell, root string, strict bool, out, errout io.Writer) int {
	sel, status := loadSelection(args, root, errout)
	if status != 0 {
		return status
	}
	var sheets []navi.Sheet
	for _, f := range sel.files {
		sheets = append(sheets, navi.Sheet{ID: sheetID(f.name), SourceName: f.name, Path: f.relative, Document: f.document})
	}
	project, errors := navi.Render(sheets, navi.Options{Shell: shell})
	report := func(issues []navi.Issue, suffix string) {
		printed := map[string]bool{}
		for _, d := range issues {
			paths := sel.sourcePaths[d.Source]
			if len(paths) == 0 {
				paths = []string{"smill"} // issue without a sheet origin, such as the profile
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
	report(errors, "")
	if len(errors) > 0 {
		return 1
	}
	report(project.Warnings, " (warning)")
	if strict && len(project.Warnings) > 0 {
		fmt.Fprintf(errout, "smill:1:1: NAV000: --strict: %d warning(s) reported; nothing was written\n", len(project.Warnings))
		return 1
	}
	if len(project.Files) == 0 {
		fmt.Fprintln(errout, "smill:1:1: NAV106: no sheet produced a .cheat file")
		return 1
	}
	if err := writeFiles(destination, project.Files); err != nil {
		fmt.Fprintf(errout, "%s:1:1: IO003: %v\n", destination, err)
		return 2
	}
	if _, err := fmt.Fprintf(out, "Exported %d sheets to %s (navi format %d, %s profile; navi was not run).\n", len(project.Files), destination, navi.Version, shell); err != nil {
		fmt.Fprintln(errout, "IO002: "+err.Error())
		return 2
	}
	return 0
}
