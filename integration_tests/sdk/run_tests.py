import argparse
import os
import subprocess
import sys


def _execute_command(args, cwd=None) -> None:
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


def _run_tests(dir_name: str, concurrency: int, rerun_failed: bool) -> None:
    if rerun_failed:
        _execute_command(["pytest", dir_name, "-lf", "-rP", "-vv", "-n", str(concurrency)])
    else:
        _execute_command(["pytest", dir_name, "-rP", "-vv", "-n", str(concurrency)])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--data-integration",
        dest="data_integration_tests",
        default=False,
        action="store_true",
        help="Run the SDK Data Integration tests.",
    )

    parser.add_argument(
        "--aqueduct",
        dest="aqueduct_tests",
        default=False,
        action="store_true",
        help="Run the SDK Aqueduct tests.",
    )

    parser.add_argument(
        "-lf",
        dest="rerun_failed",
        default=False,
        action="store_true",
        help="Run only the tests in the suite that failed during the last run.",
    )

    parser.add_argument(
        "-n",
        dest="concurrency",
        default=8,
        action="store",
        help="The concurrency to run the test suite with.",
    )

    args = parser.parse_args()

    if not (args.aqueduct_tests or args.data_integration_tests):
        args.aqueduct_tests = True
        args.data_integration_tests = True

    cwd = os.getcwd()
    if not cwd.endswith("integration_tests/sdk"):
        print("Current directory should be the SDK integration test directory.")
        print("Your working directory is %s" % cwd)
        exit(1)

    if args.aqueduct_tests:
        print("Running Aqueduct Tests...")
        _run_tests("aqueduct_tests/", args.concurrency, args.rerun_failed)

    if args.data_integration_tests:
        print("Running Data Integration Tests...")
        _run_tests("data_integration_tests/", args.concurrency, args.rerun_failed)