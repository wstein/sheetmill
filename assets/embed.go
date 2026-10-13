// SPDX-License-Identifier: EUPL-1.2
package brand

import "embed"

// Files contains local web assets only; font/generation tools are not needed at runtime.
//
//go:embed logo-accessible.svg logo-dark.svg favicon.svg favicon-16.png favicon-32.png apple-touch-icon.png fonts/IBMPlexSans-Regular.woff2 licenses/IBMPlexSans-web-OFL.txt tokens.json
var Files embed.FS
