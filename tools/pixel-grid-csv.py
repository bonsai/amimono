#!/usr/bin/env python3
"""Convert amimono pixel-grid JSON to/from a simple CSV matrix.

CSV cells contain palette IDs. Palette metadata remains in the JSON source.
This keeps the CSV useful as a universal interchange format while preserving
colors and semantics in the canonical PixelGrid document.
"""

import argparse
import csv
import json
from pathlib import Path


def json_to_csv(src: Path, dst: Path) -> None:
    data = json.loads(src.read_text(encoding="utf-8"))
    cells = data["cells"]
    width = data["width"]
    height = data["height"]
    if len(cells) != height or any(len(row) != width for row in cells):
        raise ValueError("cells dimensions do not match width/height")
    with dst.open("w", encoding="utf-8", newline="") as f:
        csv.writer(f).writerows(cells)


def csv_to_json(src: Path, dst: Path, name: str) -> None:
    with src.open("r", encoding="utf-8", newline="") as f:
        cells = [[int(v) for v in row] for row in csv.reader(f)]
    if not cells or not cells[0]:
        raise ValueError("empty CSV")
    width = len(cells[0])
    if any(len(row) != width for row in cells):
        raise ValueError("CSV rows have inconsistent widths")
    data = {
        "schema": "amimono/pixel-grid/v1",
        "name": name,
        "width": width,
        "height": len(cells),
        "coordinate": {
            "origin": "top-left",
            "rows": "1-based top-to-bottom",
            "columns": "1-based left-to-right",
        },
        "cells": cells,
    }
    dst.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


parser = argparse.ArgumentParser()
sub = parser.add_subparsers(dest="command", required=True)

p1 = sub.add_parser("to-csv")
p1.add_argument("src", type=Path)
p1.add_argument("dst", type=Path)

p2 = sub.add_parser("from-csv")
p2.add_argument("src", type=Path)
p2.add_argument("dst", type=Path)
p2.add_argument("--name", default="pixel grid")

args = parser.parse_args()
if args.command == "to-csv":
    json_to_csv(args.src, args.dst)
else:
    csv_to_json(args.src, args.dst, args.name)
