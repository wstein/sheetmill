// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wstein/sheetmill/internal/markdown"
)

const Version = "dev"
const usage = "usage: smill version | smill check <file-or-directory>... | smill inspect <file> | smill import reference <file> [options]"

func Run(args []string, out, errout io.Writer) int {
	if len(args) > 0 && args[0] == "import" {
		return runImport(args[1:], out, errout)
	}
	fail := func(message string) int { fmt.Fprintln(errout, message); return 2 }
	if len(args) == 1 && args[0] == "version" {
		if _, err := fmt.Fprintln(out, "smill "+Version); err != nil {
			return fail("IO002: " + err.Error())
		}
		return 0
	}
	if len(args) < 2 || (args[0] != "check" && args[0] != "inspect") || (args[0] == "inspect" && len(args) != 2) {
		return fail(usage)
	}
	var files []string
	seen := map[string]bool{}
	status := 0
	for _, p := range args[1:] {
		paths, err := inputs(p, args[0] == "check")
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", p, err)
			status = 2
			continue
		}
		for _, p := range paths {
			if !seen[p] {
				seen[p] = true
				files = append(files, p)
			}
		}
	}
	sort.Strings(files)
	for _, p := range files {
		b, err := read(p)
		if err != nil {
			fmt.Fprintf(errout, "%s:1:1: IO001: %v\n", p, err)
			status = 2
			continue
		}
		d, ds := markdown.Parse(b)
		if len(ds) > 0 {
			for _, diag := range ds {
				fmt.Fprintf(errout, "%s:%s\n", p, diag.Error())
			}
			status = max(status, 1)
			continue
		}
		if args[0] == "inspect" {
			b, err := d.JSON()
			if err != nil {
				return fail("IO002: " + err.Error())
			}
			if _, err = out.Write(b); err != nil {
				return fail("IO002: " + err.Error())
			}
		}
	}
	return status
}

// os.Root confines each open beneath its explicitly selected parent. Symlinks,
// including ancestor symlinks, are rejected before opening. Root prevents escape
// if a child is replaced between the checks and the read.
func checkedPath(path string) (string, os.FileInfo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	current := string(filepath.Separator)
	var info os.FileInfo
	for _, part := range strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err = os.Lstat(current)
		if err != nil {
			return "", nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", nil, fmt.Errorf("symlink paths are unsupported: %s", current)
		}
	}
	return abs, info, nil
}
func inputs(path string, allowDir bool) ([]string, error) {
	abs, info, err := checkedPath(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		if !allowDir {
			return nil, fmt.Errorf("inspect requires a regular .md file")
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}
		var paths []string
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") {
				if entry.Type()&os.ModeSymlink != 0 {
					return nil, fmt.Errorf("symlink input is unsupported: %s", entry.Name())
				}
				if !entry.IsDir() {
					paths = append(paths, filepath.Join(abs, entry.Name()))
				}
			}
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("directory contains no .md files (nonrecursive)")
		}
		return paths, nil
	}
	if !info.Mode().IsRegular() || !strings.HasSuffix(abs, ".md") {
		return nil, fmt.Errorf("expected a regular .md file")
	}
	return []string{abs}, nil
}
func read(path string) ([]byte, error) {
	abs, info, err := checkedPath(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("expected a regular file")
	}
	root, err := os.OpenRoot(filepath.Dir(abs))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	file, err := root.Open(filepath.Base(abs))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := io.ReadAll(io.LimitReader(file, markdown.MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	return b, nil
}
