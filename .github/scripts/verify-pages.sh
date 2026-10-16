#!/bin/sh
# SPDX-License-Identifier: EUPL-1.2
#
# Verify the assembled GitHub Pages tree (specs/018-layout-paths, FR-7): the
# SheetMill manual at /sheetmill/ plus the same demo library in three layouts at
# /sheetmill/demo/, /sheetmill/demo-compact/ and /sheetmill/demo-cards/.
#
# Usage: sh .github/scripts/verify-pages.sh PUBLIC_DIR
#
# Reads the local tree only; no network, no writes. Tools: sh, grep -E, sed, find,
# sort. Exit 0 when every rule holds, 1 when a rule fails, 2 on a usage error.
# The generator writes double-quoted attributes; the tag-anchored check (rule 0)
# fails on anything else, so the rules below may rely on that.

set -u

if [ "$#" -ne 1 ] || [ ! -d "$1" ]; then
	echo "usage: sh verify-pages.sh PUBLIC_DIR" >&2
	exit 2
fi

root=${1%/}
origin=https://wstein.github.io/sheetmill
base=/sheetmill
home_editorial=$base/demo/
home_compact=$base/demo-compact/
home_cards=$base/demo-cards/
bad=0

fail() {
	printf 'FAIL rule %s: %s\n' "$1" "$2" >&2
	bad=1
}

# html_files SITE_DIR lists the HTML pages that belong to a site. The manual owns
# everything except the three demo directories, so a stray demo* directory is
# still checked as manual content.
html_files() {
	if [ "$1" = "$root" ]; then
		find "$root" -type f -name '*.html' ! -path "$root/demo/*" \
			! -path "$root/demo-compact/*" ! -path "$root/demo-cards/*" | sort
	else
		find "$1" -type f -name '*.html' | sort
	fi
}

# grep_all SITE_DIR PATTERN prints "FILE MATCH" for every match in the site pages.
# File names come from the generator and contain no spaces or "|".
grep_all() {
	html_files "$1" | while IFS= read -r f; do
		grep -oE "$2" "$f" | sed "s|^|$f |"
	done
}

# attr_values SITE_DIR PREFIX prints "FILE VALUE" for every root-relative
# href, src or action value.
attr_values() {
	grep_all "$1" ' (href|src|action)="/[^"]*"' | sed -E 's/^([^ ]*) +(href|src|action)="([^"]*)"$/\1 \3/'
}

# ---- Rule 0: attributes are double-quoted (tag-anchored) -------------------------
for site in "$root" "$root/demo" "$root/demo-compact" "$root/demo-cards"; do
	[ -d "$site" ] || continue
	hits=$(grep_all "$site" '<[a-zA-Z][^>]* (href|src|action)=[^"]' | sed -E 's/ .*$//' | sort -u)
	[ -z "$hits" ] || fail 0 "single-quoted or unquoted href/src/action in: $(echo "$hits" | tr '\n' ' ')"
done

# ---- Rule 1: top-level entries ---------------------------------------------------
for name in demo demo-compact demo-cards assets brand; do
	[ -d "$root/$name" ] || fail 1 "missing directory $name"
done
for name in llms.txt agent-guide.md index.html sitemap.xml search.json; do
	[ -f "$root/$name" ] || fail 1 "missing file $name"
done
for entry in $(find "$root" -maxdepth 1 -mindepth 1 | sed -E 's|^.*/||' | sort); do
	case "$entry" in
	demo | demo-compact | demo-cards) ;;
	demo*) fail 1 "reserved prefix demo used by top-level entry $entry" ;;
	esac
done

# ---- Rule 2: every site is complete ----------------------------------------------
for site in "$root" "$root/demo" "$root/demo-compact" "$root/demo-cards"; do
	[ -d "$site" ] || continue
	for name in index.html sitemap.xml search.json; do
		[ -f "$site/$name" ] || fail 2 "${site#"$root"} lacks $name"
	done
	[ -d "$site/assets" ] || fail 2 "${site#"$root"} lacks assets/"
done

# ---- Rule 3: root-relative URLs stay inside their own base path ------------------
# check_prefix SITE_DIR BASE: a demo may only use its own base or a layout home.
check_prefix() {
	vals=$(attr_values "$1")
	while IFS=' ' read -r f v; do
		[ -n "$f" ] || continue
		case "$v" in
		//*) continue ;; # protocol-relative: rule 8
		"$2"*) continue ;;
		"$home_editorial" | "$home_compact" | "$home_cards") continue ;;
		esac
		fail 3 "$f: $v is outside $2"
	done <<EOF
$vals
EOF
}
for pair in "demo:$home_editorial" "demo-compact:$home_compact" "demo-cards:$home_cards"; do
	site=$root/${pair%%:*}
	[ -d "$site" ] && check_prefix "$site" "${pair#*:}"
done
# The manual may name a demo only by a layout home or a path below one.
vals=$(attr_values "$root")
while IFS=' ' read -r f v; do
	[ -n "$f" ] || continue
	case "$v" in
	//*) continue ;;
	"$home_editorial"* | "$home_compact"* | "$home_cards"*) continue ;;
	"$base"/demo*) fail 3 "$f: $v is not a layout home or below one" ;;
	"$base"/*) continue ;;
	*) fail 3 "$f: $v is outside $base/" ;;
	esac
done <<EOF
$vals
EOF

# ---- Rule 4: absolute URLs, sitemaps, placeholder --------------------------------
# check_absolute SITE_DIR SITE_URL: canonical, og:url, og:image and sitemap <loc>.
check_absolute() {
	vals=$(grep_all "$1" 'rel="canonical" href="[^"]*"|property="og:(url|image)" content="[^"]*"' |
		sed -E 's/^([^ ]*) .*="([^"]*)"$/\1 \2/')
	if [ -f "$1/sitemap.xml" ]; then
		locs=$(grep -oE '<loc>[^<]*</loc>' "$1/sitemap.xml" | sed -E 's|<loc>([^<]*)</loc>|\1|')
		vals="$vals
$(printf '%s\n' "$locs" | sed -E "s|^|$1/sitemap.xml |")"
	fi
	while IFS=' ' read -r f v; do
		[ -n "$f" ] || continue
		if [ "$1" = "$root" ]; then
			case "$v" in
			"$origin/demo"*) fail 4 "$f: manual URL $v is below a demo" ;;
			"$origin/"*) ;;
			*) fail 4 "$f: manual URL $v is outside $origin/" ;;
			esac
		else
			case "$v" in
			"$2"*) ;;
			*) fail 4 "$f: $v does not begin with $2" ;;
			esac
		fi
	done <<EOF
$vals
EOF
}
check_absolute "$root" "$origin/"
for name in demo demo-compact demo-cards; do
	[ -d "$root/$name" ] && check_absolute "$root/$name" "$origin/$name/"
done
leaks=$(find "$root" -type f -exec grep -l 'example\.invalid' {} + 2>/dev/null)
[ -z "$leaks" ] || fail 4 "example.invalid placeholder leaked into: $(echo "$leaks" | tr '\n' ' ')"

# ---- Rule 5: layout markers ------------------------------------------------------
# check_markers SITE_DIR NAME WANT_CLASSIC WANT_CARDS reports every page whose
# body class tokens or stylesheets disagree with the layout.
check_markers() {
	[ -d "$1" ] || return 0
	for f in $(html_files "$1"); do
		body=$(grep -oE '<body[^>]*>' "$f" | head -n 1)
		classes=$(printf '%s\n' "$body" | sed -nE 's/.*class="([^"]*)".*/\1/p')
		has_classic=0
		has_cards=0
		css_classic=0
		css_cards=0
		case " $classes " in *" classic "*) has_classic=1 ;; esac
		case " $classes " in *" reference-cards "*) has_cards=1 ;; esac
		grep -qE 'assets/classic\.[0-9a-f]{64}\.css' "$f" && css_classic=1
		grep -qE 'assets/cards\.[0-9a-f]{64}\.css' "$f" && css_cards=1
		if [ "$has_classic$css_classic$has_cards$css_cards" != "$3$3$4$4" ]; then
			fail 5 "$f: $2 expects classic=$3 cards=$4 but has body-classic=$has_classic css-classic=$css_classic body-cards=$has_cards css-cards=$css_cards"
		fi
	done
}
check_markers "$root" manual 0 0
check_markers "$root/demo" editorial 0 0
check_markers "$root/demo-compact" compact 1 0
check_markers "$root/demo-cards" cards 1 1

# ---- Rule 6: the demos have the same pages ---------------------------------------
pages_of() {
	html_files "$1" | sed -E "s|^$1/||" | sort
}
if [ -d "$root/demo" ]; then
	ref=$(pages_of "$root/demo")
	for name in demo-compact demo-cards; do
		[ -d "$root/$name" ] || continue
		other=$(pages_of "$root/$name")
		if [ "$ref" != "$other" ]; then
			diffs=
			for p in $ref; do
				case "
$other
" in *"
$p
"*) ;; *) diffs="$diffs -$p" ;; esac
			done
			for p in $other; do
				case "
$ref
" in *"
$p
"*) ;; *) diffs="$diffs +$p" ;; esac
			done
			fail 6 "$name page set differs from demo:$diffs"
		fi
	done
fi

# ---- Rule 7: header links name exactly the other two layouts ---------------------
# header_hrefs FILE prints the href of every header-link element, one per line.
header_hrefs() {
	flat=$(sed -E ':a;N;$!ba;s/\n/ /g' "$1")
	{
		printf '%s\n' "$flat" | sed -nE 's/.*(<nav class="header-links".*)/\1/p' | sed -E 's|</nav>.*||' | grep -oE 'href="[^"]*"'
		printf '%s\n' "$flat" | sed -nE 's/.*(<p class="hero-links">.*)/\1/p' | sed -E 's|</p>.*||' | grep -oE 'href="[^"]*"'
		printf '%s\n' "$flat" | grep -oE '<a class="topbar-link" href="[^"]*">[^<]*' | grep -vE '>Library$' | sed -E 's/<a class="topbar-link" (href="[^"]*").*/\1/'
	} | sed -E 's/^href="(.*)"$/\1/' | sort -u
}
check_links() {
	[ -f "$1/index.html" ] || return 0
	got=$(header_hrefs "$1/index.html")
	want=$(printf '%s\n%s\n' "$3" "$4" | sort)
	if [ "$got" != "$want" ]; then
		fail 7 "$2 home header links are [$(echo "$got" | tr '\n' ' ')], want [$(echo "$want" | tr '\n' ' ')]"
	fi
}
check_links "$root/demo" editorial "$home_compact" "$home_cards"
check_links "$root/demo-compact" compact "$home_editorial" "$home_cards"
check_links "$root/demo-cards" cards "$home_editorial" "$home_compact"

# ---- Rule 8: no protocol-relative or insecure subresource URLs -------------------
for site in "$root" "$root/demo" "$root/demo-compact" "$root/demo-cards"; do
	[ -d "$site" ] || continue
	hits=$(grep_all "$site" ' (href|src|action)="//[^"]*"')
	[ -z "$hits" ] || fail 8 "protocol-relative URL: $(echo "$hits" | head -n 3 | tr '\n' ' ')"
	# Plain <a href="http:..."> prose links are exempt (documents show loopback examples).
	hits=$(grep_all "$site" '<link [^>]*href="http:[^"]*"|<[a-zA-Z]+ [^>]*(src|action)="http:[^"]*"')
	[ -z "$hits" ] || fail 8 "http: subresource, link or action: $(echo "$hits" | head -n 3 | tr '\n' ' ')"
done

if [ "$bad" -ne 0 ]; then
	echo "verify-pages: FAILED" >&2
	exit 1
fi
echo "verify-pages: ok"
