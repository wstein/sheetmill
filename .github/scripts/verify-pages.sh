#!/bin/sh
# SPDX-License-Identifier: EUPL-1.2
#
# Verify the assembled GitHub Pages tree (specs/018-layout-paths, FR-7): the
# SheetMill manual at /sheetmill/ plus the same demo library in three layouts at
# /sheetmill/demo/, /sheetmill/demo-compact/ and /sheetmill/demo-cards/.
#
# Usage: sh .github/scripts/verify-pages.sh PUBLIC_DIR
#
# Reads the local tree only; no network and no writes (no temporary files: lists
# are shell variables split on newlines, not here-documents or process
# substitution). Tools: sh, grep -E, find, sort, tr. Exit 0 when every rule holds,
# 1 when a rule fails or a tool errors, 2 on a usage error. PUBLIC_DIR may contain
# only [A-Za-z0-9/._-]. The generator writes one tag per line with double-quoted
# attributes; the tag-anchored check (rule 0) fails on anything else, so the rules
# below may rely on that.

set -u
set -f
NL='
'
IFS=$NL

if [ "$#" -ne 1 ] || [ ! -d "$1" ]; then
	echo "usage: sh verify-pages.sh PUBLIC_DIR" >&2
	exit 2
fi
case $1 in
*[!A-Za-z0-9/._-]*)
	echo "verify-pages: PUBLIC_DIR may contain only [A-Za-z0-9/._-]" >&2
	exit 2
	;;
esac

root=${1%/}
origin=https://wstein.github.io/sheetmill
base=/sheetmill
home_editorial=$base/demo/
home_compact=$base/demo-compact/
home_cards=$base/demo-cards/
bad=0

# fail RULE MESSAGE records a violation; RULE "tool" marks a tool or read error.
fail() {
	printf 'FAIL rule %s: %s\n' "$1" "$2" >&2
	bad=1
}

# list_pages SITE_DIR sets PAGES to the HTML pages of a site, one per line. The
# manual owns everything except the three demo directories, so a stray demo*
# directory is still checked as manual content.
list_pages() {
	if [ "$1" = "$root" ]; then
		lp_out=$(find "$root" -type f -name '*.html' ! -path "$root/demo/*" \
			! -path "$root/demo-compact/*" ! -path "$root/demo-cards/*")
	else
		lp_out=$(find "$1" -type f -name '*.html')
	fi || fail tool "cannot list $1"
	PAGES=$(printf '%s\n' "$lp_out" | sort)
}

# each_match SITE_DIR GREP_FLAGS PATTERN CALLBACK calls CALLBACK FILE MATCH for
# every match of PATTERN in the site pages. A grep error (status above 1) fails.
each_match() {
	list_pages "$1"
	for em_file in $PAGES; do
		[ -n "$em_file" ] || continue
		em_out=$(grep "$2" "$3" "$em_file")
		if [ "$?" -gt 1 ]; then
			fail tool "cannot read $em_file"
			continue
		fi
		for em_match in $em_out; do
			"$4" "$em_file" "$em_match"
		done
	done
}

# has FILE PATTERN: 0 if FILE matches, 1 if not; a grep error fails the run.
has() {
	grep -qE "$2" "$1"
	case $? in
	0) return 0 ;;
	1) return 1 ;;
	*)
		fail tool "cannot read $1"
		return 1
		;;
	esac
}

# attr_value MATCH prints the text of the last double-quoted attribute value.
attr_value() {
	av=${1##*=\"}
	printf '%s' "${av%\"}"
}

sites=$root$NL$root/demo$NL$root/demo-compact$NL$root/demo-cards

# ---- Rule 0: attributes are double-quoted (tag-anchored) -------------------------
cb_quote() { fail 0 "$1: single-quoted or unquoted href/src/action in $2"; }
for site in $sites; do
	[ -d "$site" ] || continue
	each_match "$site" -ioE '<[a-zA-Z][^>]*[[:space:]](href|src|action)=([^"]|$)' cb_quote
done

# ---- Rule 1: top-level entries ---------------------------------------------------
for name in demo demo-compact demo-cards assets brand; do
	[ -d "$root/$name" ] || fail 1 "missing directory $name"
done
for name in llms.txt agent-guide.md index.html sitemap.xml search.json; do
	[ -f "$root/$name" ] || fail 1 "missing file $name"
done
entries=$(find "$root" -maxdepth 1 -mindepth 1) || fail tool "cannot list $root"
for entry in $entries; do
	case "${entry##*/}" in
	demo | demo-compact | demo-cards) ;;
	demo*) fail 1 "reserved prefix demo used by top-level entry ${entry##*/}" ;;
	esac
done

# ---- Rule 2: every site is complete ----------------------------------------------
for site in $sites; do
	[ -d "$site" ] || continue
	for name in index.html sitemap.xml search.json; do
		[ -f "$site/$name" ] || fail 2 "${site#"$root"} lacks $name"
	done
	[ -d "$site/assets" ] || fail 2 "${site#"$root"} lacks assets/"
done

# ---- Rule 3: root-relative URLs stay inside their own base path ------------------
# A demo may only use its own base or a layout home; the manual may name a demo
# only by a layout home or a path below one.
r3_base=
r3_kind=
cb_prefix() {
	v=$(attr_value "$2")
	case "$v" in
	//*) return ;; # protocol-relative: rule 8
	"$home_editorial" | "$home_compact" | "$home_cards") return ;;
	esac
	if [ "$r3_kind" = demo ]; then
		case "$v" in "$r3_base"*) return ;; esac
		fail 3 "$1: $v is outside $r3_base"
	else
		case "$v" in
		"$home_editorial"* | "$home_compact"* | "$home_cards"*) ;;
		"$base"/demo*) fail 3 "$1: $v is not a layout home or below one" ;;
		"$base"/*) ;;
		*) fail 3 "$1: $v is outside $base/" ;;
		esac
	fi
}
for pair in "demo:$home_editorial" "demo-compact:$home_compact" "demo-cards:$home_cards"; do
	[ -d "$root/${pair%%:*}" ] || continue
	r3_base=${pair#*:}
	r3_kind=demo
	each_match "$root/${pair%%:*}" -ioE '[[:space:]](href|src|action)="/[^"]*"' cb_prefix
done
r3_kind=manual
each_match "$root" -ioE '[[:space:]](href|src|action)="/[^"]*"' cb_prefix

# ---- Rule 4: absolute URLs, canonical links, sitemaps, placeholder ---------------
# A demo's absolute URLs begin with its own site URL; the manual's with the project
# origin and never below a demo.
r4_url=
cb_absolute() {
	v=$(attr_value "$2")
	r4_check "$1" "$v"
}
cb_loc() {
	v=${2#<loc>}
	r4_check "$1" "${v%</loc>}"
}
r4_check() {
	if [ "$r4_url" = "$origin/" ]; then
		case "$2" in
		"$origin"/demo*) fail 4 "$1: manual URL $2 is below a demo" ;;
		"$origin/"*) ;;
		*) fail 4 "$1: manual URL $2 is outside $origin/" ;;
		esac
	else
		case "$2" in
		"$r4_url"*) ;;
		*) fail 4 "$1: $2 does not begin with $r4_url" ;;
		esac
	fi
}
cb_count() { r4_count=$((r4_count + 1)); }
for site in $sites; do
	[ -d "$site" ] || continue
	if [ "$site" = "$root" ]; then r4_url=$origin/; else r4_url=$origin/${site##*/}/; fi
	each_match "$site" -oE 'rel="canonical" href="[^"]*"|property="og:(url|image)" content="[^"]*"' cb_absolute
	# Exactly one canonical link on every page.
	list_pages "$site"
	for page in $PAGES; do
		[ -n "$page" ] || continue
		r4_count=0
		r4_out=$(grep -oE '<link[^>]*rel="canonical"' "$page")
		[ "$?" -le 1 ] || fail tool "cannot read $page"
		for r4_m in $r4_out; do cb_count; done
		[ "$r4_count" -eq 1 ] || fail 4 "$page has $r4_count canonical links, want 1"
	done
	if [ -f "$site/sitemap.xml" ]; then
		locs=$(grep -oE '<loc>[^<]*</loc>' "$site/sitemap.xml")
		[ "$?" -le 1 ] || fail tool "cannot read $site/sitemap.xml"
		for loc in $locs; do cb_loc "$site/sitemap.xml" "$loc"; done
	fi
done
leaks=$(find "$root" -type f -exec grep -l 'example\.invalid' {} +)
[ "$?" -le 1 ] || fail tool "cannot search $root"
[ -z "$leaks" ] || fail 4 "example.invalid placeholder leaked into: $(printf '%s' "$leaks" | tr '\n' ' ')"

# ---- Rule 5: layout markers ------------------------------------------------------
# check_markers SITE_DIR NAME WANT_CLASSIC WANT_CARDS reports every page whose
# body class tokens or stylesheets disagree with the layout.
check_markers() {
	[ -d "$1" ] || return 0
	list_pages "$1"
	for f in $PAGES; do
		[ -n "$f" ] || continue
		body=$(grep -oE '<body[^>]*>' "$f")
		[ "$?" -le 1 ] || fail tool "cannot read $f"
		body=${body%%"$NL"*}
		[ -n "$body" ] || fail 5 "$f has no body element"
		classes=
		case "$body" in
		*' class="'*)
			classes=${body#*' class="'}
			classes=${classes%%\"*}
			;;
		esac
		has_classic=0
		has_cards=0
		css_classic=0
		css_cards=0
		case " $classes " in *" classic "*) has_classic=1 ;; esac
		case " $classes " in *" reference-cards "*) has_cards=1 ;; esac
		has "$f" 'assets/classic\.[0-9a-f]{64}\.css' && css_classic=1
		has "$f" 'assets/cards\.[0-9a-f]{64}\.css' && css_cards=1
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
# relative_pages SITE_DIR sets REL to its HTML pages relative to the site, one per line.
relative_pages() {
	list_pages "$1"
	REL=
	for rp in $PAGES; do
		[ -n "$rp" ] || continue
		REL=$REL${rp#"$1"/}$NL
	done
}
# missing_from LIST ITEMS prints every item of ITEMS that LIST lacks.
missing_from() {
	mf_out=
	for mf in $2; do
		case "$NL$1" in *"$NL$mf$NL"*) ;; *) mf_out="$mf_out $mf" ;; esac
	done
	printf '%s' "$mf_out"
}
if [ -d "$root/demo" ]; then
	relative_pages "$root/demo"
	ref=$REL
	for name in demo-compact demo-cards; do
		[ -d "$root/$name" ] || continue
		relative_pages "$root/$name"
		lacks=$(missing_from "$REL" "$ref")
		extra=$(missing_from "$ref" "$REL")
		if [ -n "$lacks$extra" ]; then
			fail 6 "$name page set differs from demo; missing:${lacks:- none}; extra:${extra:- none}"
		fi
	done
fi

# ---- Rule 7: header links name exactly the other two layouts ---------------------
# check_kind PAGE KIND REQUIRED HREFS WANT1 WANT2: the hrefs of one header-link
# element kind are exactly WANT1 and WANT2, each once; REQUIRED=1 demands the kind.
check_kind() {
	ck_n=0
	ck_a=0
	ck_b=0
	for ck in $4; do
		ck_n=$((ck_n + 1))
		[ "$ck" = "$5" ] && ck_a=$((ck_a + 1))
		[ "$ck" = "$6" ] && ck_b=$((ck_b + 1))
	done
	if [ "$ck_n" -eq 0 ] && [ "$3" -eq 0 ]; then return; fi
	if [ "$ck_n" -ne 2 ] || [ "$ck_a" -ne 1 ] || [ "$ck_b" -ne 1 ]; then
		fail 7 "$1: $2 links are [$(printf '%s' "$4" | tr '\n' ' ')], want exactly $5 and $6 once each"
	fi
}
# element_hrefs FLAT START END prints the hrefs between START and the next END.
element_hrefs() {
	case "$1" in
	*"$2"*) ;;
	*) return 0 ;;
	esac
	eh=${1#*"$2"}
	eh=${eh%%"$3"*}
	printf '%s\n' "$eh" | grep -oE 'href="[^"]*"'
	[ "$?" -le 1 ] || fail tool "cannot read header links"
}
check_links() {
	[ -f "$1/index.html" ] || return 0
	flat=$(tr '\n' ' ' <"$1/index.html")
	nav=
	hero=
	top=
	nav=$(element_hrefs "$flat" '<nav class="header-links"' '</nav>')
	hero=$(element_hrefs "$flat" '<p class="hero-links">' '</p>')
	top=$(printf '%s\n' "$flat" | grep -oE '<a class="topbar-link" href="[^"]*">[^<]*')
	[ "$?" -le 1 ] || fail tool "cannot read $1/index.html"
	topurls=
	for t in $top; do
		case "$t" in *'>Library') continue ;; esac
		t=${t#*href=\"}
		topurls=$topurls${t%%\"*}$NL
	done
	strip() {
		st=
		for s in $1; do
			s=${s#href=\"}
			st=$st${s%\"}$NL
		done
		printf '%s' "$st"
	}
	nav=$(strip "$nav")
	hero=$(strip "$hero")
	if [ "$2" = editorial ]; then nav_req=1; top_req=0; else nav_req=0; top_req=1; fi
	check_kind "$1/index.html" "nav.header-links" "$nav_req" "$nav" "$3" "$4"
	check_kind "$1/index.html" "p.hero-links" 1 "$hero" "$3" "$4"
	check_kind "$1/index.html" "a.topbar-link" "$top_req" "$topurls" "$3" "$4"
}
check_links "$root/demo" editorial "$home_compact" "$home_cards"
check_links "$root/demo-compact" compact "$home_editorial" "$home_cards"
check_links "$root/demo-cards" cards "$home_editorial" "$home_compact"
# The manual's Example library link keeps reaching the editorial demo.
if [ -f "$root/index.html" ]; then
	has "$root/index.html" "href=\"$home_editorial\"" || fail 7 "manual home does not link $home_editorial"
fi

# ---- Rule 8: no protocol-relative or foreign subresource URLs --------------------
# Plain <a href="http:..."> prose links are exempt (documents show loopback examples
# such as http://127.0.0.1:1313/). Any other absolute src, srcset, action or
# <link href> must stay on the project origin; <base> is not used.
cb_protorel() { fail 8 "$1: protocol-relative URL in $2"; }
cb_foreign() {
	v=$(attr_value "$2")
	case "$v" in
	"$origin/"*) ;;
	*) fail 8 "$1: absolute subresource, link or action $v is not under $origin/" ;;
	esac
}
cb_base() { fail 8 "$1: unexpected <base> element"; }
for site in $sites; do
	[ -d "$site" ] || continue
	each_match "$site" -ioE '[[:space:]](href|src|srcset|action)="//[^"]*"' cb_protorel
	each_match "$site" -ioE '<[a-z][a-z0-9]*[^>]*[[:space:]](src|srcset|action)="[a-z][a-z0-9+.-]*:[^"]*"' cb_foreign
	each_match "$site" -ioE '<link[^>]*[[:space:]]href="[a-z][a-z0-9+.-]*:[^"]*"' cb_foreign
	each_match "$site" -ioE '<base([[:space:]/>]|$)' cb_base
done

if [ "$bad" -ne 0 ]; then
	echo "verify-pages: FAILED" >&2
	exit 1
fi
echo "verify-pages: ok"
