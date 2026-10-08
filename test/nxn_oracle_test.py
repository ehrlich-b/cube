import unittest

from nxn_oracle import check_move_mean, check_two_by_two_latency, parse_documented_metrics


class MoveRegressionContract(unittest.TestCase):
    def test_exactly_five_percent_is_allowed(self):
        check_move_mean(4, "uniform", [124, 128], 120)

    def test_exceeding_five_percent_fails(self):
        with self.assertRaisesRegex(AssertionError, "regressed"):
            check_move_mean(4, "uniform", [126, 128], 120)

    def test_limit_applies_to_mean_not_maximum(self):
        check_move_mean(4, "uniform", [80, 160], 120)

    def test_documented_table_requires_every_size_once(self):
        rows = [f"| {n}×{n} | 100.00 / 120 | 99.00 / 119 | 50.00 / 80.00 ms |"
                for n in (2, 4, 5, 6, 7)]
        parsed = parse_documented_metrics("\n".join(rows), "fixture")
        self.assertEqual(parsed[4], (100, 120, 99, 119, 50, 80))
        for broken in (rows[:-1], rows[:-1] + [rows[0]]):
            with self.assertRaises(AssertionError):
                parse_documented_metrics("\n".join(broken), "fixture")


class TwoByTwoLatencyContract(unittest.TestCase):
    def test_budget_allows_exact_boundary(self):
        self.assertEqual(check_two_by_two_latency([100, 140], [150]), (120, 150))

    def test_mean_regression_fails(self):
        with self.assertRaisesRegex(AssertionError, "fresh-process mean regressed"):
            check_two_by_two_latency([100, 142], [150])

    def test_busy_machine_scales_budget_from_reference(self):
        self.assertEqual(check_two_by_two_latency([250, 350], [400]), (300, 400))
        with self.assertRaises(AssertionError):
            check_two_by_two_latency([400], [400])

    def test_isolated_slow_process_does_not_gate_maximum(self):
        check_two_by_two_latency([20, 20, 300], [150])
