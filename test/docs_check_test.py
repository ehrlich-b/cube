"""Regression tests for the executable-docs checker (stdlib only)."""

import os
from pathlib import Path
import tempfile
import unittest

import docs_check as docs


class MarkdownTests(unittest.TestCase):
    def test_multiline_quotes_and_inline_expectation(self):
        examples = docs.extract('''```sh
make build
cd cube
./dist/cube verify "R U U' R'" \\
  --target "YB|Y9/R9/B9/W9/O9/G9"
cube optimize "R R R" # → R'
```''', "example.md")
        self.assertEqual(len(examples), 1)
        self.assertEqual(examples[0].commands, [
            (["verify", "R U U' R'", "--target", "YB|Y9/R9/B9/W9/O9/G9"], None),
            (["optimize", "R R R"], "R'"),
        ])

    def test_expected_block_attaches_to_previous_example(self):
        examples = docs.extract('''```bash
cube optimize "R R"
```
Expected output:
```text cube-output contains
Optimized: R2 (1 moves)
```''', "example.md")
        self.assertEqual(examples[0].expected, "Optimized: R2 (1 moves)")
        self.assertTrue(examples[0].contains)

    def test_unmarked_workflow_and_orphan_output_fail(self):
        with self.assertRaisesRegex(ValueError, "cube-check"):
            docs.extract('```sh\nsolution=$(cube solve R)\n```', "bad.md")
        with self.assertRaisesRegex(ValueError, "cube-check"):
            docs.extract('```sh\nCUBE_CACHE_DIR=.scratch/cache cube solve R\n```', "bad.md")
        with self.assertRaisesRegex(ValueError, "must follow"):
            docs.extract('```text cube-output\nsolved\n```', "bad.md")

    def test_output_contradictions_and_order_fail(self):
        with self.assertRaisesRegex(ValueError, "stdout differs"):
            docs.compare("Optimized: R\n", "Optimized: R2\n")
        with self.assertRaisesRegex(ValueError, "out-of-order"):
            docs.compare("second\nfirst\n", "first\nsecond\n", contains=True)
        docs.compare("banner\nfirst\nextra\nsecond\n", "first\nsecond", contains=True)


class ExecutionTests(unittest.TestCase):
    def setUp(self):
        (docs.ROOT / ".scratch").mkdir(exist_ok=True)
        self.temp = tempfile.TemporaryDirectory(prefix="docs-check-test-", dir=docs.ROOT / ".scratch")
        self.addCleanup(self.temp.cleanup)
        self.binary = Path(self.temp.name) / "cube"
        self.binary.write_text('''#!/bin/sh
case "$1" in
  fail) echo rejected >&2; exit 7 ;;
  empty-lookup) echo 'No algorithms found.' ;;
  *) printf '%s\\n' "$*" ;;
esac
''')
        self.binary.chmod(0o700)

    def test_nonzero_cli_and_legacy_false_success_fail(self):
        with self.assertRaisesRegex(ValueError, "exit 7"):
            docs.run_example(docs.Example("test", commands=[(["fail"], None)]), self.binary, os.environ, 5)
        with self.assertRaisesRegex(ValueError, "despite exit 0"):
            docs.run_example(docs.Example("test", commands=[(["empty-lookup"], None)]), self.binary, os.environ, 5)

    def test_workflow_checks_substitution_and_expected_output(self):
        example = docs.Example("test", script='moves=$(./dist/cube first)\ncube "$moves" second', expected="first second")
        docs.run_example(example, self.binary, os.environ, 5)
        example.expected = "wrong"
        with self.assertRaisesRegex(ValueError, "stdout differs"):
            docs.run_example(example, self.binary, os.environ, 5)

    def test_masked_nonzero_in_pipeline_still_fails(self):
        example = docs.Example("test", script='moves=$(cube fail | cat) || true\ncube ok')
        with self.assertRaisesRegex(ValueError, "cube exited 7"):
            docs.run_example(example, self.binary, os.environ, 5)


if __name__ == "__main__":
    unittest.main()
