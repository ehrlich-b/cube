#!/usr/bin/env python3
"""Run Markdown CLI examples against dist/cube, without installing anything."""

import argparse
from dataclasses import dataclass, field
import difflib
import os
from pathlib import Path
import re
import shlex
import signal
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
CUBE = re.compile(r"(?<![\w./-])(?:\./dist/)?cube(?=\s|$)")
SHELLS = {"sh", "bash", "shell"}


@dataclass
class Example:
    location: str
    commands: list = field(default_factory=list)
    script: str = ""
    expected: str | None = None
    contains: bool = False


def extract(text, source):
    examples = []
    pending = None
    fence = None
    lines = []
    for number, line in enumerate(text.splitlines(), 1):
        if fence is None:
            match = re.match(r"^```(.*)$", line)
            if match:
                fence = (match[1].split(), number)
                lines = []
            continue
        if line != "```":
            lines.append(line)
            continue
        info, start = fence
        fence = None
        body = "\n".join(lines)
        location = f"{source}:{start + 1}"
        if info and info[0] in SHELLS:
            example = Example(location)
            if "cube-check" in info:
                if not CUBE.search(body):
                    raise ValueError(f"{location}: cube-check block has no cube command")
                example.script = body
            else:
                # Only CLI lines run: never clone, build, install, or publish
                # commands that happen to share the README's shell fence.
                for command in re.sub(r"\\\n\s*", " ", body).splitlines():
                    if not CUBE.match(command.lstrip()):
                        nested = re.search(r"(?:\$\(|&&|\|\||[|;])\s*(?:\./dist/)?cube\b", command)
                        prefixed = any(command[match.end():].strip() for match in CUBE.finditer(command))
                        if (nested or (prefixed and not command.lstrip().startswith("cd "))) and not command.lstrip().startswith("#"):
                            raise ValueError(f"{location}: mark shell workflows with cube-check")
                        continue
                    argv = shlex.split(command, comments=True)
                    if any(token in {"|", "||", "&&", ";", ">", "<"} or "$" in token for token in argv):
                        raise ValueError(f"{location}: mark shell workflows with cube-check")
                    arrow = re.search(r"#\s*→\s*(.*)$", command)
                    example.commands.append((argv[1:], arrow[1].strip() if arrow else None))
            if example.commands or example.script:
                examples.append(example)
                pending = example
            else:
                pending = None
        elif info[:2] == ["text", "cube-output"]:
            if not pending or pending.expected is not None:
                raise ValueError(f"{location}: output block must follow a runnable example")
            if info[2:] not in ([], ["contains"]):
                raise ValueError(f"{location}: unsupported output comparison")
            pending.expected = body
            pending.contains = info[2:] == ["contains"]
            pending = None
        else:
            pending = None
    if fence:
        raise ValueError(f"{source}:{fence[1]}: unclosed code fence")
    return examples


def normalize(output):
    output = re.sub(r"\x1b\[[0-9;]*m", "", output)
    return "\n".join(line.rstrip() for line in output.splitlines()).strip("\n")


def compare(actual, expected, contains=False):
    actual, expected = normalize(actual), normalize(expected)
    if contains:
        remaining = iter(actual.splitlines())
        for line in expected.splitlines():
            if not any(candidate == line for candidate in remaining):
                raise ValueError(f"missing or out-of-order expected line: {line!r}")
    elif actual != expected:
        diff = "\n".join(difflib.unified_diff(expected.splitlines(), actual.splitlines(), fromfile="expected", tofile="actual", lineterm=""))
        raise ValueError(f"stdout differs:\n{diff}")


def run_process(argv, env, timeout):
    process = subprocess.Popen(argv, cwd=ROOT, env=env, stdin=subprocess.PIPE,
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                               text=True, start_new_session=True)
    try:
        stdout, stderr = process.communicate("quit\n", timeout=timeout)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid, signal.SIGKILL)
        process.communicate()
        raise ValueError(f"command exceeded {timeout}s: {shlex.join(argv)}") from None
    if process.returncode:
        raise ValueError(f"exit {process.returncode}: {shlex.join(argv)}\n{stdout}{stderr}")
    if re.search(r"(?m)^(?:No algorithms found\.|Error:|❌ (?:FAIL|NO MATCH))", stdout):
        raise ValueError(f"command reported failure despite exit 0:\n{stdout}")
    return stdout


def run_example(example, binary, env, timeout):
    if example.script:
        # Record each CLI invocation, even inside substitutions/pipelines or
        # where shell syntax masks a nonzero exit with `|| true`.
        with tempfile.TemporaryDirectory(prefix="docs-check-", dir=ROOT / ".scratch") as scratch:
            receipts = Path(scratch) / "failures"
            workflow_env = dict(env, CUBE_DOCS_BIN=str(binary), CUBE_DOCS_FAILURES=str(receipts))
            wrapper = '''cube() {
  if "$CUBE_DOCS_BIN" "$@"; then return 0; else
    local cube_status=$?
    printf 'cube exited %s: %s\\n' "$cube_status" "$*" >> "$CUBE_DOCS_FAILURES"
    return "$cube_status"
  fi
}
'''
            output = run_process(["bash", "-euo", "pipefail", "-c", wrapper + CUBE.sub("cube", example.script)], workflow_env, timeout)
            if receipts.exists():
                raise ValueError(receipts.read_text())
    else:
        outputs = []
        for argv, arrow in example.commands:
            output = run_process([str(binary), *argv], env, timeout)
            if arrow and arrow not in output:
                raise ValueError(f"expected inline output {arrow!r}, got:\n{output}")
            outputs.append(output)
        output = "".join(outputs)
    if example.expected is not None:
        compare(output, example.expected, example.contains)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, default=ROOT / "dist/cube")
    parser.add_argument("--timeout", type=float, default=90)
    args = parser.parse_args()
    binary = args.binary.resolve()
    if not binary.is_file():
        parser.error(f"binary not found: {binary}; run make build first")
    scratch = ROOT / ".scratch"
    scratch.mkdir(exist_ok=True)
    env = dict(os.environ, TMPDIR=str(scratch), TMP=str(scratch), TEMP=str(scratch))
    env.setdefault("CUBE_CACHE_DIR", str(scratch / "cube-cache"))
    files = [ROOT / "README.md", *sorted((ROOT / "examples").glob("*.md"))]
    failures = []
    count = 0
    passed = 0
    for file in files:
        try:
            examples = extract(file.read_text(), file.relative_to(ROOT))
        except ValueError as error:
            failures.append(str(error))
            continue
        for example in examples:
            count += 1
            try:
                run_example(example, binary, env, args.timeout)
                passed += 1
            except ValueError as error:
                failures.append(f"{example.location}: {error}")
    for failure in failures:
        print(f"FAIL {failure}")
    print(f"Docs: {passed}/{count} runnable blocks passed across {len(files)} files")
    return bool(failures)


if __name__ == "__main__":
    raise SystemExit(main())
