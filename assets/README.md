# Editorial Stack source candidates

The user selected a three-sheet stack with the coral/periwinkle palette on
2026-10-13. [BRAND.md](BRAND.md) governs the direction. Exact SVG geometry remains
pending human visual approval; [preview.svg](preview.svg) compares the candidates.

`icon.svg` is the source mark; `icon-dark.svg` adds the defined dark tile.
`logo.svg` and `logo-dark.svg` use outlined Bricolage Grotesque in one shaped run,
with a shared baseline and natural kerning across Sheet/Mill. The light
`logo-accessible.svg` uses dark coral for its wordmark accent. No external fonts,
images, scripts or linked resources are needed to display any SVG.

Only these five logo/icon files and the comparison board currently exist. Small
optical masters, favicon derivatives, banners, social cards and renderer samples
remain future work. The preview's tiny icons are normal-master samples, not a
claim that the separate small-size master has been approved.

## Reproduction inputs

- Python 3.15.0 was used for this revision.
- `requirements.txt` pins FontTools 4.66.1 and uharfbuzz 0.56.3.
- Bricolage Grotesque source: [Google Fonts at 66928797](https://github.com/google/fonts/blob/66928797f271fed5ee428544e7229685e2d90127/ofl/bricolagegrotesque/BricolageGrotesque%5Bopsz%2Cwdth%2Cwght%5D.ttf).
- SHA-256: `413e7357809ddd12fd80a96a8a396de0e401638d4acd3cb3e37532f0472ac682`.
- Font axes: weight 700, width 100, optical size 72. The source font is kept
  outside the repository; its [copyright/OFL notice](licenses/BricolageGrotesque-OFL.txt)
  is retained here. Original SheetMill drawing/generator uses EUPL-1.2.

Obtain that exact font separately, then use an isolated tool environment:

```sh
uv venv --python 3.15 /private/tmp/sheetmill-brand-venv
uv pip install --python /private/tmp/sheetmill-brand-venv/bin/python -r assets/requirements.txt
/private/tmp/sheetmill-brand-venv/bin/python assets/generate.py --font /absolute/path/BricolageGrotesque.ttf
```

The generator downloads nothing and fails on a mismatched font hash or tool
version. Never use an implicit host font fallback. `tokens.json` records the
candidate geometry and color inputs; changes must agree with `BRAND.md`.

## Verification performed

- The six SVG outputs reproduced byte-for-byte on a second generation.
- Parsed every SVG as XML; each has role, title and description, with no live
  text, scripts, external images, foreign objects, uses or resource links.
- Rendered the comparison board with Inkscape 1.4.4 (`dcaf3e7`, 2026-05-05) and
  inspected light/dark wordmarks, the darker accent and 128/64/32/16 px samples.
- Rendered the icon at 1024 px and estimated filled-area shares using alpha-weighted
  nearest-palette classification: front coral 63.90%, pale middle 18.05%, rear
  periwinkle 18.05%. This meets the proposed 60% coral dominance target; it is a
  raster approximation, not a vector-area proof.
- Checked the contrast pairs recorded in `BRAND.md`. Display-wordmark colors
  rely on the specific logo exception, not a general text exemption.

Local PNG previews live outside the repository. No production raster exports,
PDFs, website, release, renderer or deployment were produced. Final geometry,
monochrome identity approval and small-size optical-master work remain open.
