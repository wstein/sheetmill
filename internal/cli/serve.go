// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wstein/sheetmill/internal/studio"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func serveCommand(out, errout io.Writer, status *int) *cobra.Command {
	var options studio.Options
	var open bool
	cmd := &cobra.Command{Use: "serve [file-or-directory]", Short: "Preview saved files locally (read-only, installed Hugo required)", Args: cobra.MaximumNArgs(1), Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			options.Input = args[0]
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		session, err := studio.Start(ctx, options)
		if err != nil {
			fmt.Fprintln(errout, "STU002: "+err.Error())
			*status = 2
			return
		}
		defer func() {
			if err := session.Close(); err != nil {
				fmt.Fprintln(errout, "STU002: preview cleanup failed")
				*status = 2
			}
		}()
		if _, err = fmt.Fprintf(out, "SheetMill Studio — read-only\nSelected: %s\nOpen: %s\n", session.Scope(), session.BootstrapURL()); err != nil {
			fmt.Fprintln(errout, "IO002: cannot print preview address")
			*status = 2
			return
		}
		if open {
			if err = openBrowser(ctx, session.BootstrapURL()); err != nil {
				fmt.Fprintln(errout, "Browser launch failed; use the printed local address.")
			}
		}
		if err = session.Wait(); err != nil {
			fmt.Fprintln(errout, "STU002: "+err.Error())
			*status = 2
		}
	}}
	cmd.Flags().BoolVar(&open, "open", false, "open the local browser after readiness")
	cmd.Flags().IntVar(&options.Port, "port", 0, "IPv4 loopback port (0 chooses an available port)")
	cmd.Flags().StringVar(&options.Hugo, "hugo", "", "absolute installed Hugo 0.167.0 executable (default: hugo on PATH)")
	return cmd
}
func openBrowser(ctx context.Context, url string) error {
	tool := "xdg-open"
	args := []string{url}
	if runtime.GOOS == "darwin" {
		tool = "/usr/bin/open"
	} else if runtime.GOOS == "windows" {
		tool = "rundll32.exe"
		args = []string{"url.dll,FileProtocolHandler", url}
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, tool, args...).Run()
}
