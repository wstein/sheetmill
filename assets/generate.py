#!/usr/bin/env python3
# SPDX-License-Identifier: EUPL-1.2
"""Generate final SVG masters from pinned local inputs; no downloads/fallbacks."""
import argparse
import hashlib
import html
import json
from pathlib import Path

import fontTools
import uharfbuzz as hb
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont

ROOT = Path(__file__).resolve().parent


def number(value):
    return f"{value:.4f}".rstrip("0").rstrip(".") if value else "0"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--font", type=Path, required=True)
    parser.add_argument("--sans-font", type=Path, required=True)
    args = parser.parse_args()
    if fontTools.__version__ != "4.66.1" or hb.__version__ != "0.56.3":
        parser.error("use the exact versions in assets/requirements.txt")
    tokens = json.loads((ROOT / "tokens.json").read_text())
    raw = args.font.read_bytes()
    if hashlib.sha256(raw).hexdigest() != tokens["wordmark"]["fontSHA256"]:
        parser.error("font hash mismatch; no substitution is allowed")
    font = TTFont(args.font)
    face = hb.Face(raw)
    shaper = hb.Font(face)
    shaper.scale = (face.upem, face.upem)
    shaper.set_variations(tokens["wordmark"]["fontAxes"])
    hb.ot_font_set_funcs(shaper)
    instance = instantiateVariableFont(font, tokens["wordmark"]["fontAxes"], inplace=False)
    glyphs = instance.getGlyphSet()
    cap = instance["OS/2"].sCapHeight
    brand_set=(shaper,instance,glyphs,cap)
    sans_raw=args.sans_font.read_bytes()
    if hashlib.sha256(sans_raw).hexdigest()!=tokens["tagline"]["fontSHA256"]:
        parser.error("sans font hash mismatch; no substitution is allowed")
    sans_source=TTFont(args.sans_font)
    sans_face=hb.Face(sans_raw)
    sans_shaper=hb.Font(sans_face)
    sans_shaper.scale=(sans_face.upem,sans_face.upem)
    sans_shaper.set_variations(tokens["tagline"]["fontAxes"])
    hb.ot_font_set_funcs(sans_shaper)
    sans_instance=instantiateVariableFont(sans_source,tokens["tagline"]["fontAxes"],inplace=False)
    sans_set=(sans_shaper,sans_instance,sans_instance.getGlyphSet(),sans_instance["OS/2"].sCapHeight)
    colors = tokens["colors"]

    def lettering(text, x, baseline, scale, fill, accent=None, sans=False):
        shaper,instance,glyphs,_=sans_set if sans else brand_set
        buf = hb.Buffer()
        buf.add_str(text)
        buf.guess_segment_properties()
        hb.shape(shaper, buf, {"kern": True, "liga": True})
        cursor = 0
        paths = []
        for info, pos in zip(buf.glyph_infos, buf.glyph_positions):
            pen = SVGPathPen(glyphs, ntos=number)
            glyphs[instance.getGlyphName(info.codepoint)].draw(pen)
            color = accent if accent and info.cluster >= 5 else fill
            dx = x + (cursor + pos.x_offset) * scale
            dy = baseline - pos.y_offset * scale
            paths.append(f'<path fill="{color}" transform="translate({number(dx)} {number(dy)}) scale({number(scale)} {number(-scale)})" d="{pen.getCommands()}"/>')
            cursor += pos.x_advance
        return "\n".join(paths), cursor * scale

    def icon(dark=False, small=False, mono=False):
        pieces = []
        if dark:
            pieces.append(f'<rect width="64" height="64" rx="{tokens["icon"]["tileRadius"]}" fill="{colors["dark-surface"]}"/>')
        for layer in tokens["icon"]["smallLayers" if small else "layers"]:
            pieces.append(f'<rect x="{layer["x"]}" y="{layer["y"]}" width="{layer["width"]}" height="{layer["height"]}" rx="{layer["radius"]}" fill="{colors["ink-900"] if mono else colors[layer["color"]]}"/>')
        return "\n".join(pieces)

    def svg(title, description, width, height, content):
        return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {number(width)} {number(height)}" width="{number(width)}" height="{number(height)}" role="img" aria-labelledby="title description">
<title id="title">{html.escape(title)}</title>
<desc id="description">{html.escape(description)}</desc>
<!-- Original SheetMill artwork: EUPL-1.2. Font outlines: Bricolage Grotesque and IBM Plex Sans, SIL OFL-1.1; see licenses/. Implemented master geometry; review record in README.md. -->
{content}
</svg>
'''

    for name, dark in [("icon.svg", False), ("icon-dark.svg", True)]:
        (ROOT / name).write_text(svg("SheetMill", "Three staggered reference sheets in coral, pale coral and periwinkle.", 64, 64, icon(dark)))
    for name,small,mono in [("icon-small.svg",True,False),("favicon.svg",True,False),("icon-mono.svg",False,True)]:
        content=icon(small=small,mono=mono)
        if mono:
            # Knockout lines retain three distinct sheets in monochrome.
            content=content.replace('<rect x="15"', '<rect stroke="#FFFFFF" stroke-width="1.5" x="15"').replace('<rect x="9"','<rect stroke="#FFFFFF" stroke-width="1.5" x="9"')
        (ROOT/name).write_text(svg("SheetMill", "Editorial Stack optical or monochrome variant.",64,64,content))
    (ROOT/"apple-touch-icon.svg").write_text(svg("SheetMill", "Opaque square icon for platform masking.",64,64,f'<rect width="64" height="64" fill="{colors["dark-surface"]}"/>'+icon()))
    c = tokens["wordmark"]["capHeight"]
    scale = c / cap
    layers=tokens["icon"]["layers"]
    min_x=min(layer["x"] for layer in layers)
    min_y=min(layer["y"] for layer in layers)
    width_units=max(layer["x"]+layer["width"] for layer in layers)-min_x
    height_units=max(layer["y"]+layer["height"] for layer in layers)-min_y
    icon_scale = (1.9 * c) / height_units
    margin = 0.6 * c
    height = 1.9 * c + 2 * margin
    word_x = margin + width_units * icon_scale + 0.55 * c
    baseline = height / 2 + c / 2
    lockups = {}
    for name, dark, accessible in [("logo.svg", False, False), ("logo-dark.svg", True, False), ("logo-accessible.svg", False, True)]:
        fill = colors["white"] if dark else colors["ink-900"]
        accent = colors["coral-700"] if accessible else colors["coral-500"]
        word, width = lettering("SheetMill", word_x, baseline, scale, fill, accent)
        mark = f'<g transform="translate({number(margin-min_x*icon_scale)} {number(margin-min_y*icon_scale)}) scale({number(icon_scale)})">{icon()}</g>'
        width += word_x + margin
        lockups[name] = (mark + "\n" + word, width, height)
        (ROOT / name).write_text(svg("SheetMill", "Editorial Stack symbol beside the single joined SheetMill wordmark.", width, height, mark + "\n" + word))

    for name,width,height,logo_scale,x,y,dark in [("banner.svg",1000,220,1.4,35,24,False),("social-preview.svg",1280,640,2.3,180,150,True)]:
        content,_,_=lockups["logo-dark.svg" if dark else "logo.svg"]
        background=colors["dark-surface"] if dark else colors["mist-50"]
        tagline,advance=lettering("One source. Every reference.", x+word_x*logo_scale, y+baseline*logo_scale+42*logo_scale, (0.43*c*logo_scale)/sans_set[3],colors["dark-text"] if dark else colors["slate-600"],sans=True)
        (ROOT/name).write_text(svg("SheetMill — One source. Every reference.","Editorial Stack identity with an outlined IBM Plex Sans tagline.",width,height,f'<rect width="{width}" height="{height}" fill="{background}"/><g transform="translate({x} {y}) scale({logo_scale})">{content}</g>'+tagline))
    board = [f'<rect width="1200" height="650" fill="{colors["mist-50"]}"/>']
    def label(text, x, y, fill="ink-900", size=22):
        paths, _ = lettering(text, x, y, size / cap, colors[fill])
        board.append(paths)
    label("SheetMill / Editorial Stack", 42, 50, size=25)
    label("Final masters / review comparison", 42, 83, "slate-600", 13)
    board.append('<rect x="32" y="112" width="550" height="195" rx="16" fill="#FFFFFF"/>')
    board.append(f'<rect x="602" y="112" width="566" height="195" rx="16" fill="{colors["dark-surface"]}"/>')
    for name, x in [("logo.svg", 55), ("logo-dark.svg", 625)]:
        content, _, _ = lockups[name]
        board.append(f'<g transform="translate({x} 153)">{content}</g>')
    label("Light / display wordmark", 55, 141, "slate-600", 13)
    label("Dark / display wordmark", 625, 141, "dark-muted", 13)
    label("Icon at 128 / 64 / 32 / 16 px", 42, 350, size=18)
    for size, x in [(128, 42), (64, 205), (32, 310), (16, 385)]:
        board.append(f'<g transform="translate({x} 375) scale({size/64})">{icon(small=size<32)}</g>')
    label("Readable accent variant / coral-700", 480, 350, size=18)
    content, _, _ = lockups["logo-accessible.svg"]
    board.append(f'<g transform="translate(480 383)">{content}</g>')
    for name, x in [("coral-500",42),("periwinkle-500",245),("ink-900",448),("coral-700",651),("periwinkle-700",854)]:
        board.append(f'<rect x="{x}" y="548" width="170" height="32" rx="6" fill="{colors[name]}"/>')
        label(name, x, 607, size=12)
        label(colors[name], x, 632, "slate-600", 12)
    (ROOT / "preview.svg").write_text(svg("SheetMill brand masters", "Light and dark outlined wordmarks, icon size samples and updated coral/periwinkle palette. Master geometry is subject to the recorded review.", 1200, 650, "\n".join(board)))


if __name__ == "__main__":
    main()
