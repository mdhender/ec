#!/usr/bin/env python3
"""Extract sections from baseline.adoc into Markdown files for Hugo."""

import re
import sys
from pathlib import Path

CONTENT_DIR = Path(__file__).resolve().parent.parent / "content" / "history" / "user-manual-1978"
BASELINE = CONTENT_DIR / "baseline.adoc"

# Map each == section title to its target .md file and front matter title.
# Order matches the foo file listing.
SECTION_MAP = [
    ("Foreword",                        "foreword.md",                          "Foreword",                         10),
    ("INTRODUCTION",                    "introduction.md",                      "Introduction",                     20),
    ("HISTORY OF THE EMPYREAN CLUSTER", "history-of-the-empyrean-cluster.md",   "History of the Empyrean Cluster",  30),
    ("GAME SET UP",                     "game-set-up.md",                       "Game Set Up",                      40),
    ("BASIC UNITS",                     "basic-units.md",                       "Basic Units",                      50),
    ("COLONIES AND SHIPS",              "colonies-and-ships.md",                "Colonies and Ships",               60),
    ("MANUFACTURING",                   "manufacturing.md",                     "Manufacturing",                    70),
    ("TECHNOLOGICAL ADVANCEMENT",       "technological-advancement.md",         "Technological Advancement",        80),
    ("TRADE",                           "trade.md",                             "Trade",                            90),
    ("EXPLORATION",                     "exploration.md",                       "Exploration",                      100),
    ("REBELLION",                       "rebellion.md",                         "Rebellion",                        110),
    ("COMBAT",                          "combat.md",                            "Combat",                           120),
    ("ESPIONAGE",                       "espionage.md",                         "Espionage",                        130),
    ("CONTROL OF PLANETS",              "control-of-planets.md",                "Control of Planets",               140),
    ("COMMUNICATION",                   "communication.md",                     "Communication",                    150),
    ("VICTORY CONDITIONS",              "victory-conditions.md",                "Victory Conditions",               160),
    ("SEQUENCE OF TURN EXECUTION",      "sequence-of-turn-execution.md",        "Sequence of Turn Execution",       170),
    ("WRITING ORDERS",                  "writing-orders.md",                    "Writing Orders",                   180),
    ("I",                               "appendix-i.md",                        "Appendix I",                       190),
    ("II",                              "appendix-ii.md",                       "Appendix II",                      200),
    ("Examples",                        "examples.md",                          "Examples",                         210),
    ("Sample Documents",                "sample-documents.md",                  "Sample Documents",                 220),
    ("Colophon",                        "colophon.md",                          "Colophon",                         230),
]


def split_sections(lines: list[str]) -> dict[str, list[str]]:
    """Split the AsciiDoc file into sections keyed by the == heading text."""
    sections = {}
    current_key = None
    current_lines = []

    for line in lines:
        # Match level-2 headings: == Title
        m = re.match(r'^== (.+)$', line)
        if m:
            if current_key is not None:
                sections[current_key] = current_lines
            current_key = m.group(1).strip()
            current_lines = []
        else:
            if current_key is not None:
                current_lines.append(line)

    if current_key is not None:
        sections[current_key] = current_lines

    return sections


def convert_section(lines: list[str]) -> str:
    """Convert AsciiDoc body lines to Markdown."""
    # --- Phase 1: Collect footnotes ---
    footnote_defs = {}  # fn-id -> text
    fn_counter = [0]

    def get_fn_num(fn_id):
        if fn_id not in footnote_defs:
            fn_counter[0] += 1
            footnote_defs[fn_id] = fn_counter[0]
        return footnote_defs[fn_id]

    # Pre-scan for footnote definitions
    fn_def_pattern = re.compile(r'^:fn-([^:]+):\s*footnote:\[(.+)\]\s*$')
    fn_texts = {}
    for line in lines:
        m = fn_def_pattern.match(line)
        if m:
            fn_id = f"fn-{m.group(1)}"
            fn_texts[fn_id] = m.group(2)
            get_fn_num(fn_id)

    # --- Phase 2: Convert line by line ---
    out = []
    i = 0
    in_table = False
    table_rows = []
    table_header_done = False
    block_title = None  # holds .Title text

    while i < len(lines):
        line = lines[i]

        # Skip AsciiDoc directives
        if fn_def_pattern.match(line):
            i += 1
            continue
        if re.match(r'^:(sectnums|sectnums!):', line):
            i += 1
            continue
        if re.match(r'^\[#[^\]]+\]$', line):
            i += 1
            continue
        if re.match(r'^\[cols="[^"]*"\]$', line):
            i += 1
            continue
        if re.match(r'^\[appendix\]$', line):
            i += 1
            continue

        # Block title: .Something (but not ordered list items like ". item")
        # A block title starts with . followed by a non-space, non-. character
        m = re.match(r'^\.([A-Z][\w].*)$', line)
        if m and not in_table:
            block_title = m.group(1).strip()
            i += 1
            continue

        # Source code blocks
        if re.match(r'^\[source,\w+\]$', line):
            i += 1
            if i < len(lines) and lines[i].strip() == '----':
                i += 1
                code_lines = []
                while i < len(lines) and lines[i].strip() != '----':
                    code_lines.append(lines[i])
                    i += 1
                i += 1  # skip closing ----
                if block_title:
                    out.append(f'**{block_title}**')
                    out.append('')
                    block_title = None
                out.append('```text')
                out.extend(code_lines)
                out.append('```')
                out.append('')
            continue

        # Tables
        if line.strip() == '|===':
            if not in_table:
                in_table = True
                table_rows = []
                table_header_done = False
                i += 1
                continue
            else:
                # End of table - emit it
                in_table = False
                if block_title:
                    out.append(f'**{block_title}**')
                    out.append('')
                    block_title = None
                emit_table(out, table_rows, fn_texts, footnote_defs)
                out.append('')
                i += 1
                continue

        if in_table:
            # Accumulate table content
            raw = line.strip()
            # Handle "2+|" column span syntax: replace with regular cell separator
            raw = re.sub(r'(\d+)\+\|', '|', raw)
            # Handle ">|" alignment prefix: just strip the >
            raw = raw.replace('>|', '|')
            if raw.startswith('|'):
                table_rows.append(raw)
            elif raw and table_rows:
                # Continuation of previous cell
                table_rows[-1] += ' ' + raw
            i += 1
            continue

        # Headings
        m = re.match(r'^(={2,})\s+(.+)$', line)
        if m:
            level = len(m.group(1)) - 1  # === -> ##, ==== -> ###, etc.
            heading = m.group(2).strip()
            out.append('#' * level + ' ' + heading)
            out.append('')
            i += 1
            continue

        # Ordered lists with loweralpha
        if re.match(r'^\[loweralpha(,start=\d+)?\]$', line):
            i += 1
            letter = ord('a')
            while i < len(lines) and lines[i].startswith('. '):
                item = lines[i][2:]
                # Check for continuation lines
                while i + 1 < len(lines) and lines[i + 1] and not lines[i + 1].startswith('. ') and not lines[i + 1].startswith('=') and not lines[i + 1].startswith('[') and not lines[i + 1].startswith('|'):
                    i += 1
                    item += ' ' + lines[i]
                out.append(f'{chr(letter)}. {item}')
                letter += 1
                i += 1
            out.append('')
            continue

        # Regular ordered lists: lines starting with ". "
        if re.match(r'^\. ', line):
            num = 1
            while i < len(lines) and lines[i].startswith('. '):
                item = lines[i][2:]
                out.append(f'{num}. {item}')
                num += 1
                i += 1
            out.append('')
            continue

        # Quote blocks
        m = re.match(r'^\[quote,(.+)\]$', line)
        if m:
            author = m.group(1).strip()
            i += 1
            quote_lines = []
            while i < len(lines) and lines[i].strip():
                quote_lines.append(lines[i])
                i += 1
            for ql in quote_lines:
                out.append(f'> {ql}')
            out.append(f'>')
            out.append(f'> — {author}')
            out.append('')
            i += 1
            continue

        # Regular line - apply inline conversions
        converted = convert_inline(line, fn_texts, footnote_defs)
        if block_title and converted.strip():
            out.append(f'**{block_title}**')
            out.append('')
            block_title = None
        out.append(converted)
        i += 1

    # Append footnote definitions
    if fn_texts:
        out.append('')
        for fn_id, num in sorted(footnote_defs.items(), key=lambda x: x[1]):
            if fn_id in fn_texts:
                out.append(f'[^{num}]: {fn_texts[fn_id]}')

    return '\n'.join(out)


def convert_inline(line: str, fn_texts: dict, fn_nums: dict) -> str:
    """Apply inline AsciiDoc -> Markdown conversions."""
    # Footnote references: {fn-xxx}
    def replace_fn(m):
        fn_id = m.group(1)
        if fn_id in fn_nums:
            return f'[^{fn_nums[fn_id]}]'
        return m.group(0)
    line = re.sub(r'\{(fn-[^}]+)\}', replace_fn, line)

    # Superscript: ^text^
    line = re.sub(r'\^([^^]+)\^', r'<sup>\1</sup>', line)

    # Cross-references: <<anchor>> or <<anchor,text>>
    def replace_xref(m):
        parts = m.group(1).split(',', 1)
        if len(parts) == 2:
            return parts[1].strip()
        return parts[0].replace('-', ' ').title()
    line = re.sub(r'<<([^>]+)>>', replace_xref, line)

    # AsciiDoc + line continuation (used in table cells)
    line = line.replace(' +\n', '<br>')
    # Sometimes just " +" at end of line
    if line.endswith(' +'):
        line = line[:-2] + '<br>'

    return line


def emit_table(out: list[str], rows: list[str], fn_texts: dict, fn_nums: dict):
    """Convert collected AsciiDoc table rows to Markdown pipe table.

    AsciiDoc tables have each cell on its own line starting with |.
    The header row is the first line. We determine column count from it,
    then group subsequent |lines into rows of that many cells.
    """
    if not rows:
        return

    # Split all rows into individual cells
    all_cells = []
    for row in rows:
        # A row like "|A|B|C" splits into ['', 'A', 'B', 'C']
        # A row like "|A" splits into ['', 'A']
        parts = row.split('|')
        parts = parts[1:]  # drop leading empty before first |
        for p in parts:
            all_cells.append(p.strip())

    if not all_cells:
        return

    # Determine column count from header.
    # The header is the first row in the table which has all columns on one line.
    first_row = rows[0]
    header_cells = first_row.split('|')[1:]
    num_cols = len(header_cells)

    if num_cols == 0:
        return

    # Group cells into rows of num_cols
    table_rows = []
    for i in range(0, len(all_cells), num_cols):
        row = all_cells[i:i + num_cols]
        if len(row) < num_cols:
            row.extend([''] * (num_cols - len(row)))
        table_rows.append(row[:num_cols])

    if not table_rows:
        return

    # Apply inline conversions to all cells
    for i, row in enumerate(table_rows):
        for j, cell in enumerate(row):
            table_rows[i][j] = convert_inline(cell, fn_texts, fn_nums)

    # First row is header
    header = table_rows[0]
    out.append('| ' + ' | '.join(header) + ' |')
    out.append('| ' + ' | '.join(['---'] * len(header)) + ' |')

    for row in table_rows[1:]:
        out.append('| ' + ' | '.join(row) + ' |')


def main():
    if not BASELINE.exists():
        print(f"Error: {BASELINE} not found", file=sys.stderr)
        sys.exit(1)

    text = BASELINE.read_text(encoding='utf-8')
    lines = text.split('\n')

    sections = split_sections(lines)

    print(f"Found {len(sections)} sections in baseline.adoc:")
    for key in sections:
        print(f"  - '{key}' ({len(sections[key])} lines)")
    print()

    written = 0
    skipped = 0
    for section_title, filename, front_title, weight in SECTION_MAP:
        if section_title not in sections:
            print(f"WARNING: Section '{section_title}' not found in baseline.adoc, skipping {filename}")
            skipped += 1
            continue

        body = convert_section(sections[section_title])

        # Strip leading/trailing blank lines from body
        body = body.strip()

        front_matter = f"---\ntitle: {front_title}\nweight: {weight}\n---\n"
        content = front_matter + '\n' + body + '\n'

        outpath = CONTENT_DIR / filename
        outpath.write_text(content, encoding='utf-8')
        print(f"Wrote {filename} ({len(body)} chars)")
        written += 1

    print(f"\nDone: {written} files written, {skipped} skipped.")


if __name__ == '__main__':
    main()
