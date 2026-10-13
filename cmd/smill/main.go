// SPDX-License-Identifier: EUPL-1.2
package main

import (
	"github.com/wstein/sheetmill/internal/cli"
	"os"
)

func main() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr)) }
