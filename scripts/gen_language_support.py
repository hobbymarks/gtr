#!/usr/bin/env python3
"""Regenerate internal/lang/data/language_support.json from translate-shell LanguageData.awk."""
import json
import os
import re
import sys

def main() -> int:
    root = os.path.join(os.path.dirname(__file__), "..")
    awk = os.path.join(
        root,
        "..",
        "translate-shell",
        "include",
        "LanguageData.awk",
    )
    if len(sys.argv) > 1:
        awk = sys.argv[1]
    if not os.path.isfile(awk):
        print("missing", awk, file=sys.stderr)
        return 1

    line_re = re.compile(r'Locale\["([^"]+)"\]\["([^"]+)"\]\s*=\s*"([^"]*)"')
    locales: dict[str, dict[str, str]] = {}
    with open(awk, encoding="utf-8") as f:
        for line in f:
            m = line_re.search(line)
            if not m:
                continue
            code, attr, val = m.group(1), m.group(2), m.group(3)
            locales.setdefault(code, {})[attr] = val

    support: dict[str, dict[str, bool]] = {}
    meta: dict[str, dict[str, str]] = {}
    for code, attrs in locales.items():
        sup = attrs.get("supported-by", "")
        parts = [p.strip().lower() for p in sup.split(";") if p.strip()]
        support[code] = {"google": "google" in parts, "bing": "bing" in parts}
        meta[code] = {}
        for field in ("name", "endonym", "family", "script", "iso", "spoken-in"):
            if field in attrs:
                meta[code][field] = attrs[field]

    aliases: dict[str, str] = {}

    def add_alias(a: str, target: str) -> None:
        a = a.strip()
        if not a:
            return
        aliases[a] = target
        aliases[a.lower()] = target

    for code, attrs in locales.items():
        if "iso" in attrs:
            add_alias(attrs["iso"], code)
        for k in ("name", "name2", "endonym", "endonym2"):
            if k in attrs:
                add_alias(attrs[k].lower(), code)

    alias_line = re.compile(r'LocaleAlias\["([^"]+)"\]\s*=\s*"([^"]+)"')
    with open(awk, encoding="utf-8") as f:
        for line in f:
            m = alias_line.search(line)
            if m:
                add_alias(m.group(1), m.group(2))

    out_path = os.path.join(root, "internal", "lang", "data", "language_support.json")
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    doc = {"support": support, "aliases": aliases, "meta": meta}
    with open(out_path, "w", encoding="utf-8") as o:
        json.dump(doc, o, ensure_ascii=False, separators=(",", ":"))
    print("wrote", out_path, "locales", len(support), "aliases", len(aliases))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
