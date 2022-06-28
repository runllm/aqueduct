import argparse
import base64
import sys

from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.param_executor.spec import ParamSpec, parse_spec
from aqueduct_executor.operators.utils.logging import (
    Error,
    ExecutionLogs,
    Logs,
    TIP_UNKNOWN_ERROR,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage


def run(spec: ParamSpec) -> None:
    """
    Executes a parameter operator by storing the parameter value in the output content path.
    """
    storage = parse_storage(spec.storage_config)
    exec_logs = ExecutionLogs(user_logs=Logs())
    try:
        utils.write_artifact(
            storage,
            enums.OutputArtifactType.JSON,
            spec.output_content_path,
            spec.output_metadata_path,
            spec.val,
            system_metadata={},
        )
        exec_logs.code = enums.ExecutionCode.SUCCEEDED
        utils.write_logs(storage, spec.metadata_path, exec_logs)
    except Exception as e:
        exec_logs.code = enums.ExecutionCode.FAILED
        exec_logs.failure_reason = enums.FailureReason.SYSTEM
        exec_logs.error = Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR)
        print(f"Failed with system error. Full Logs:\n{exec_logs.json()}")
        utils.write_logs(storage, spec.metadata_path, exec_logs)
        sys.exit(1)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    print("Job Spec: \n{}".format(spec.json()))
    run(spec)
