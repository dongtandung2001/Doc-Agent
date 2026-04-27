import argparse
import json
import re
from pathlib import Path
from dataclasses import dataclass

import torch
from transformers import AutoTokenizer, AutoModelForCausalLM
from jsonschema import validate


# =========================
# LLM CONFIG + WRAPPER
# =========================


@dataclass
class LLMConfig:
    model_id: str = "codellama/CodeLlama-7b-Instruct-hf"
    max_context_tokens: int = 8192
    max_new_tokens: int = 800
    temperature: float = 0.2
    top_p: float = 0.95


class TransformersLLM:
    def __init__(self, cfg: LLMConfig):
        self.cfg = cfg
        self.tokenizer = AutoTokenizer.from_pretrained(cfg.model_id)
        self.model = AutoModelForCausalLM.from_pretrained(
            cfg.model_id,
            device_map="auto",
            torch_dtype=torch.float16 if torch.cuda.is_available() else torch.float32,
        )
        self.model.eval()

    def generate(self, instruction: str) -> str:
        prompt = f"<s>[INST] {instruction} [/INST]"
        inputs = self.tokenizer(
            prompt,
            return_tensors="pt",
            truncation=True,
            max_length=self.cfg.max_context_tokens,
        )
        inputs = {k: v.to(self.model.device) for k, v in inputs.items()}

        with torch.inference_mode():
            out = self.model.generate(
                **inputs,
                max_new_tokens=self.cfg.max_new_tokens,
                do_sample=(self.cfg.temperature > 0),
                temperature=self.cfg.temperature,
                top_p=self.cfg.top_p,
            )

        return self.tokenizer.decode(out[0], skip_special_tokens=True)


# =========================
# UTILS (merged)
# =========================


def extract_first_json_object(text: str):
    match = re.search(r"\{.*\}", text, re.DOTALL)
    if not match:
        raise ValueError("No JSON object found")
    return json.loads(match.group(0))


def extract_best_matching_json(text: str, required_keys):
    candidates = re.findall(r"\{.*?\}", text, re.DOTALL)
    for c in candidates:
        try:
            obj = json.loads(c)
            if all(k in obj for k in required_keys):
                return obj
        except:
            continue
    raise ValueError("No valid JSON found")


# =========================
# QA MODE
# =========================

QA_SYSTEM = "Return JSON only. Use only provided code."


def run_qa(args):
    code = Path(args.path).read_text(encoding="utf-8", errors="ignore")

    llm = TransformersLLM(LLMConfig())
    prompt = f"""{QA_SYSTEM}

Question: {args.question}

CODE:
{code}
"""

    raw = llm.generate(prompt)
    obj = extract_first_json_object(raw)

    schema = json.loads(Path(args.schema).read_text())
    validate(instance=obj, schema=schema)

    print(json.dumps(obj, indent=2))


# =========================
# SUMMARIZER MODE
# =========================

SUMMARY_SYSTEM = """You are a code documentation generator.

CRITICAL OUTPUT RULES:
- Output JSON ONLY.
- Output a SINGLE JSON object.
- Use ONLY specified keys.
- No placeholders.
"""


def build_prompt(code: str, file_path: str) -> str:
    return f"""{SUMMARY_SYSTEM}

Summarize file into JSON with keys:
file, overview, imports, globals, classes, functions, entrypoints, risks

file_path = {file_path}

CODE:
{code}
"""


def contains_placeholders(obj: dict) -> bool:
    dumped = json.dumps(obj)
    return "<" in dumped or ">" in dumped


def run_summary(args):
    file_path = str(args.path)
    code = Path(args.path).read_text(encoding="utf-8", errors="ignore")
    schema = json.loads(Path(args.schema).read_text())

    llm = TransformersLLM(LLMConfig(model_id=args.model_id))

    raw = llm.generate(build_prompt(code, file_path))

    obj = extract_best_matching_json(
        raw,
        required_keys=[
            "overview",
            "imports",
            "globals",
            "classes",
            "functions",
            "entrypoints",
            "risks",
        ],
    )

    defaults = {
        "file": file_path,
        "overview": [],
        "imports": [],
        "globals": [],
        "classes": [],
        "functions": [],
        "entrypoints": [],
        "risks": [],
    }

    for k, v in defaults.items():
        obj.setdefault(k, v)

    if contains_placeholders(obj):
        raise ValueError("Model output contains placeholders")

    validate(instance=obj, schema=schema)

    out = Path(args.path + ".summary.json")
    out.write_text(json.dumps(obj, indent=2), encoding="utf-8")
    print("Wrote", out)


# =========================
# BINARY SEARCH as an example
# =========================


def binary_search(arr, x):
    low = 0
    high = len(arr) - 1

    while low <= high:
        mid = low + (high - low) // 2

        if arr[mid] < x:
            low = mid + 1
        elif arr[mid] > x:
            high = mid - 1
        else:
            return mid
    return -1


# =========================
# MAIN ENTRYPOINT
# =========================


def main():
    parser = argparse.ArgumentParser(description="Unified LLM Code Tool")
    subparsers = parser.add_subparsers(dest="command")

    # QA command
    qa_parser = subparsers.add_parser("qa")
    qa_parser.add_argument("path")
    qa_parser.add_argument("--question", required=True)
    qa_parser.add_argument("--schema", default="schemas/qa.schema.json")

    # Summary command
    sum_parser = subparsers.add_parser("summarize")
    sum_parser.add_argument("path")
    sum_parser.add_argument("--schema", default="schemas/summary.schema.json")
    sum_parser.add_argument("--model_id", default="codellama/CodeLlama-7b-Instruct-hf")

    args = parser.parse_args()

    if args.command == "qa":
        run_qa(args)
    elif args.command == "summarize":
        run_summary(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
