import os
import re
import ast
import json
import argparse
from dataclasses import dataclass
from typing import List, Dict, Optional, Tuple

import torch
from transformers import AutoTokenizer, AutoModelForCausalLM, BitsAndBytesConfig


# Data structures
@dataclass
class Unit:
    kind: str            
    name: str
    qualname: str
    lineno: int
    end_lineno: int
    code: str


# Repo scanning helpers
DEFAULT_EXCLUDE_DIRS = {
    ".git", ".hg", ".svn", ".idea", ".vscode",
    "__pycache__", ".pytest_cache", ".mypy_cache",
    "venv", ".venv", "env", ".env",
    "node_modules", "dist", "build", ".tox",
}

PY_FILE_RE = re.compile(r".+\.py$", re.IGNORECASE)


def should_skip_dir(dirname: str, exclude_dirs: set) -> bool:
    base = os.path.basename(dirname)
    return base in exclude_dirs


def iter_py_files(repo_path: str, exclude_dirs: set) -> List[str]:
    files = []
    for root, dirs, filenames in os.walk(repo_path):
        # prune excluded directories
        dirs[:] = [d for d in dirs if not should_skip_dir(os.path.join(root, d), exclude_dirs)]
        for fn in filenames:
            if PY_FILE_RE.match(fn):
                files.append(os.path.join(root, fn))
    files.sort()
    return files


def read_text(path: str, max_bytes: int = 2_000_000) -> Optional[str]:
    try:
        with open(path, "rb") as f:
            data = f.read(max_bytes + 1)
        if len(data) > max_bytes:
            return None
        try:
            return data.decode("utf-8")
        except UnicodeDecodeError:
            return data.decode("latin-1")
    except Exception:
        return None

# AST extraction
class UnitExtractor(ast.NodeVisitor):
    def __init__(self, src: str, filename: str):
        self.src = src
        self.lines = src.splitlines(keepends=True)
        self.filename = filename
        self.units: List[Unit] = []
        self.class_stack: List[str] = []

    def _get_segment(self, node: ast.AST) -> str:
        if not (hasattr(node, "lineno") and hasattr(node, "end_lineno")):
            return ""
        start = max(1, int(node.lineno))
        end = max(start, int(node.end_lineno))
        return "".join(self.lines[start - 1:end]).rstrip() + "\n"

    def _qualname(self, name: str) -> str:
        if self.class_stack:
            return ".".join(self.class_stack + [name])
        return name

    def visit_ClassDef(self, node: ast.ClassDef):
        self.class_stack.append(node.name)
        code = self._get_segment(node)
        if code:
            self.units.append(Unit(
                kind="class",
                name=node.name,
                qualname=self._qualname(node.name),
                lineno=int(getattr(node, "lineno", 1)),
                end_lineno=int(getattr(node, "end_lineno", getattr(node, "lineno", 1))),
                code=code
            ))
        self.generic_visit(node)
        self.class_stack.pop()

    def visit_FunctionDef(self, node: ast.FunctionDef):
        code = self._get_segment(node)
        if code:
            self.units.append(Unit(
                kind="function",
                name=node.name,
                qualname=self._qualname(node.name),
                lineno=int(getattr(node, "lineno", 1)),
                end_lineno=int(getattr(node, "end_lineno", getattr(node, "lineno", 1))),
                code=code
            ))
        self.generic_visit(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef):
        code = self._get_segment(node)
        if code:
            self.units.append(Unit(
                kind="function",
                name=node.name,
                qualname=self._qualname(node.name),
                lineno=int(getattr(node, "lineno", 1)),
                end_lineno=int(getattr(node, "end_lineno", getattr(node, "lineno", 1))),
                code=code
            ))
        self.generic_visit(node)


def extract_units(src: str, filename: str) -> List[Unit]:
    try:
        tree = ast.parse(src)
    except SyntaxError:
        return []

    extractor = UnitExtractor(src, filename)
    extractor.visit(tree)
    extractor.units.sort(key=lambda u: (u.lineno, u.end_lineno))
    return extractor.units


def short_file_header(src: str, max_lines: int = 80) -> str:
    lines = src.splitlines()
    return "\n".join(lines[:max_lines]).rstrip() + "\n"


# LLM loading and generation
def load_llm(model_id: str, use_4bit: bool = True):
    if torch.cuda.is_available():
        compute_dtype = torch.float16
        device_map = {"": 0}  # force GPU 0 (prevents CPU/disk dispatch error)
    else:
        compute_dtype = torch.float32
        device_map = {"": "cpu"}

    kwargs = {}

    if use_4bit and torch.cuda.is_available():
        bnb_config = BitsAndBytesConfig(
            load_in_4bit=True,
            bnb_4bit_quant_type="nf4",
            bnb_4bit_use_double_quant=True,
            bnb_4bit_compute_dtype=compute_dtype,
        )
        kwargs["quantization_config"] = bnb_config

    tok = AutoTokenizer.from_pretrained(model_id, use_fast=True)
    if tok.pad_token is None:
        tok.pad_token = tok.eos_token
    tok.padding_side = "left"

    model = AutoModelForCausalLM.from_pretrained(
        model_id,
        device_map=device_map,
        torch_dtype=compute_dtype if torch.cuda.is_available() else None,
        **kwargs
    )
    model.eval()
    return tok, model


def generate(tok, model, prompt: str, max_new_tokens: int = 200, temperature: float = 0.0) -> str:
    inputs = tok(prompt, return_tensors="pt", truncation=True, max_length=min(tok.model_max_length, 8192))
    inputs = {k: v.to(model.device) for k, v in inputs.items()}

    with torch.no_grad():
        out = model.generate(
            **inputs,
            max_new_tokens=max_new_tokens,
            do_sample=(temperature > 0),
            temperature=temperature if temperature > 0 else 1.0,
            top_p=0.9 if temperature > 0 else 1.0,
            eos_token_id=tok.eos_token_id,
            pad_token_id=tok.pad_token_id,
        )

    text = tok.decode(out[0], skip_special_tokens=True)
    if text.startswith(prompt):
        return text[len(prompt):].strip()
    return text.strip()


# Prompt templates
def unit_prompt(lang: str, file_rel: str, unit: Unit) -> str:
    parts = []
    parts.append(f"You are a senior {lang} engineer and technical writer.")
    parts.append(f"Summarize the following {unit.kind} from a codebase in plain English.")
    parts.append("")
    parts.append("Rules:")
    parts.append("- Be concise (2–5 sentences).")
    parts.append("- Mention purpose, key inputs/outputs, and important side effects.")
    parts.append("- If it calls external services/files/db/network, mention it.")
    parts.append("- Avoid speculative claims.")
    parts.append("")
    parts.append(f"File: {file_rel}")
    parts.append(f"Symbol: {unit.qualname} (lines {unit.lineno}-{unit.end_lineno})")
    parts.append("")
    parts.append("### Code")
    parts.append("```python")
    parts.append(unit.code.rstrip())
    parts.append("```")
    parts.append("")
    parts.append("### Summary")
    return "\n".join(parts) + "\n"


def file_prompt(lang: str, file_rel: str, header_snip: str, unit_summaries: List[Tuple[str, str]]) -> str:
    parts = []
    parts.append(f"You are a senior {lang} engineer.")
    parts.append("Write a file-level summary for the code below based on (a) the file header and (b) extracted symbol summaries.")
    parts.append("")
    parts.append("Rules:")
    parts.append("- 1 short paragraph + 3–8 bullet points.")
    parts.append("- Include key responsibilities, public APIs, and any notable dependencies.")
    parts.append("- If the file seems like a script/CLI/entrypoint, say so.")
    parts.append("")
    parts.append(f"File: {file_rel}")
    parts.append("")
    parts.append("### File header (first lines)")
    parts.append("```python")
    parts.append(header_snip.rstrip())
    parts.append("```")
    parts.append("")
    parts.append("### Extracted symbol summaries")

    if unit_summaries:
        for name, summ in unit_summaries[:60]:
            parts.append(f"- **{name}**: {summ}")
    else:
        parts.append("- (No functions/classes extracted)")

    parts.append("")
    parts.append("### File summary")
    return "\n".join(parts) + "\n"


def repo_prompt(lang: str, repo_name: str, file_summaries: List[Tuple[str, str]]) -> str:
    parts = []
    parts.append(f"You are a senior {lang} engineer.")
    parts.append("Write a repository-level summary based on file summaries.")
    parts.append("")
    parts.append("Output format:")
    parts.append("1) One-paragraph overview")
    parts.append("2) Architecture/components (5–12 bullets)")
    parts.append("3) Notable flows (2–6 bullets)")
    parts.append("4) Risks/unknowns (0–6 bullets)")
    parts.append("5) Suggested next steps for a new developer (3–8 bullets)")
    parts.append("")
    parts.append(f"Repository: {repo_name}")
    parts.append("")
    parts.append("### File summaries")

    if file_summaries:
        for path, summ in file_summaries[:120]:
            parts.append(f"- **{path}**: {summ}")
    else:
        parts.append("- (No file summaries available)")

    parts.append("")
    parts.append("### Repo summary")
    return "\n".join(parts) + "\n"


# Main pipeline
def summarize_repo(
    repo_path: str,
    model_id: str,
    out_md: str,
    lang: str = "Python",
    max_files: int = 200,
    max_units_per_file: int = 60,
    max_unit_code_chars: int = 6000,
    max_new_tokens_unit: int = 180,
    max_new_tokens_file: int = 260,
    max_new_tokens_repo: int = 450,
    use_4bit: bool = True,
    temperature: float = 0.0,
):
    tok, model = load_llm(model_id, use_4bit=use_4bit)

    repo_path = os.path.abspath(repo_path)
    repo_name = os.path.basename(repo_path.rstrip("\\/"))
    py_files = iter_py_files(repo_path, DEFAULT_EXCLUDE_DIRS)[:max_files]

    file_results: List[Dict] = []
    file_summaries_for_repo: List[Tuple[str, str]] = []

    for fpath in py_files:
        rel = os.path.relpath(fpath, repo_path)
        src = read_text(fpath)
        if src is None:
            file_results.append({
                "file": rel,
                "skipped": True,
                "reason": "file too large or unreadable",
            })
            continue

        units = extract_units(src, rel)
        units = units[:max_units_per_file]

        unit_summaries: List[Tuple[str, str]] = []
        for u in units:
            code = u.code
            if len(code) > max_unit_code_chars:
                code = code[:max_unit_code_chars] + "\n# ... (truncated)\n"
                u = Unit(u.kind, u.name, u.qualname, u.lineno, u.end_lineno, code)

            up = unit_prompt(lang, rel, u)
            us = generate(tok, model, up, max_new_tokens=max_new_tokens_unit, temperature=temperature)
            unit_summaries.append((u.qualname, us))

        header_snip = short_file_header(src)
        fp = file_prompt(lang, rel, header_snip, unit_summaries)
        fs = generate(tok, model, fp, max_new_tokens=max_new_tokens_file, temperature=temperature)

        file_results.append({
            "file": rel,
            "skipped": False,
            "units": [{"name": n, "summary": s} for n, s in unit_summaries],
            "file_summary": fs,
        })
        file_summaries_for_repo.append((rel, fs))

    rp = repo_prompt(lang, repo_name, file_summaries_for_repo)
    repo_summary = generate(tok, model, rp, max_new_tokens=max_new_tokens_repo, temperature=temperature)

    with open(out_md, "w", encoding="utf-8") as f:
        f.write(f"# Repository Summary: {repo_name}\n\n")
        f.write(repo_summary.strip() + "\n\n")
        f.write("## File Summaries\n\n")

        for fr in file_results:
            f.write(f"### `{fr['file']}`\n\n")
            if fr.get("skipped"):
                f.write(f"_Skipped_: {fr.get('reason','unknown')}\n\n")
                continue

            f.write(fr["file_summary"].strip() + "\n\n")

            if fr["units"]:
                f.write("<details>\n<summary>Symbol summaries</summary>\n\n")
                for u in fr["units"]:
                    f.write(f"- **{u['name']}**: {u['summary']}\n")
                f.write("\n</details>\n\n")

    out_json = os.path.splitext(out_md)[0] + ".json"
    with open(out_json, "w", encoding="utf-8") as jf:
        json.dump({
            "repo": repo_name,
            "repo_path": repo_path,
            "model": model_id,
            "repo_summary": repo_summary,
            "files": file_results,
        }, jf, indent=2)

    print(f"Done.\n- Markdown report: {out_md}\n- JSON report: {out_json}")


def main():
    ap = argparse.ArgumentParser(description="Summarize a Python codebase using an LLM (inference-only).")
    ap.add_argument("repo", help="Path to the repository folder")
    ap.add_argument("--model", default="meta-llama/CodeLlama-7b-Instruct-hf", help="HF model id or local path")
    ap.add_argument("--out", default="repo_summary.md", help="Output markdown file")
    ap.add_argument("--no-4bit", action="store_true", help="Disable 4-bit loading (uses more VRAM)")
    ap.add_argument("--max-files", type=int, default=200, help="Max .py files to process")
    ap.add_argument("--max-units", type=int, default=60, help="Max functions/classes per file")
    ap.add_argument("--max-unit-chars", type=int, default=6000, help="Max chars per function/class code snippet")
    ap.add_argument("--t", type=float, default=0.0, help="Temperature (0 = deterministic)")
    args = ap.parse_args()

    summarize_repo(
        repo_path=args.repo,
        model_id=args.model,
        out_md=args.out,
        max_files=args.max_files,
        max_units_per_file=args.max_units,
        max_unit_code_chars=args.max_unit_chars,
        use_4bit=(not args.no_4bit),
        temperature=args.t,
    )


if __name__ == "__main__":
    main()