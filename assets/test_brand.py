# SPDX-License-Identifier: EUPL-1.2
"""Verify final brand dimensions, self-contained resources and approved colors."""
import json
import struct
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parent
NS = '{http://www.w3.org/2000/svg}'


class BrandAssets(unittest.TestCase):
    def test_svg_contract(self):
        palette = set(json.loads((ROOT / 'tokens.json').read_text())['colors'].values())
        svg_files = list(ROOT.glob('*.svg'))
        self.assertGreaterEqual(len(svg_files), 12)
        for file in svg_files:
            with self.subTest(file=file.name):
                root = ET.parse(file).getroot()
                self.assertEqual(root.get('role'), 'img')
                self.assertTrue(root.find(NS + 'title').text)
                self.assertTrue(root.find(NS + 'desc').text)
                ids = [node.get('id') for node in root.iter() if node.get('id')]
                self.assertEqual(len(ids), len(set(ids)))
                for node in root.iter():
                    self.assertNotIn(node.tag.split('}')[-1], ['text', 'script', 'image', 'foreignObject', 'use'])
                    self.assertFalse(any(key.split('}')[-1] == 'href' for key in node.attrib))
                    for key in ['fill', 'stroke']:
                        if key in node.attrib:
                            self.assertIn(node.get(key), palette)

    def test_export_dimensions(self):
        sizes = {'favicon-16.png': (16, 16), 'favicon-32.png': (32, 32),
                 'apple-touch-icon.png': (180, 180), 'banner.png': (1000, 220),
                 'social-preview.png': (1280, 640)}
        for name, size in sizes.items():
            with self.subTest(file=name):
                data = (ROOT / name).read_bytes()
                self.assertEqual(data[:8], b'\x89PNG\r\n\x1a\n')
                self.assertEqual(struct.unpack('>II', data[16:24]), size)
        for name, size in {'banner.svg': (1000, 220), 'social-preview.svg': (1280, 640)}.items():
            self.assertEqual(tuple(map(float, ET.parse(ROOT / name).getroot().get('viewBox').split()))[2:], size)

    def test_optical_and_opaque_variants(self):
        normal = ET.parse(ROOT / 'icon.svg').getroot().findall(NS + 'rect')
        small = ET.parse(ROOT / 'icon-small.svg').getroot().findall(NS + 'rect')
        self.assertEqual(len(normal), 3)
        self.assertEqual(len(small), 3)
        self.assertNotEqual([n.attrib for n in normal], [n.attrib for n in small])
        self.assertEqual((ROOT / 'favicon.svg').read_bytes(), (ROOT / 'icon-small.svg').read_bytes())
        square = ET.parse(ROOT / 'apple-touch-icon.svg').getroot().find(NS + 'rect')
        self.assertEqual(square.get('width'), '64')
        self.assertEqual(square.get('height'), '64')
        self.assertIsNone(square.get('rx'))


if __name__ == '__main__':
    unittest.main()
