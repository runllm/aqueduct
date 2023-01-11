import os

import pytest

# Maps the test files in this directory to the allowed data integrations for that file.
# If a disallowed data integration is used, all tests in the file will be skipped.
from aqueduct.constants.enums import ServiceType

allowed_data_integrations_by_file = {"relational_test": [ServiceType.SQLITE, ServiceType.SNOWFLAKE]}


@pytest.fixture(autouse=True)
def filter_tests_based_on_data_integrations(request, client, data_integration):
    """Does the same thing as `enable_only_for_data_integration_type()`, only over entire files.

    This is because the data integration tests are grouped such that each file is only relevant for
    a specific integration(s).

    All that is required is that every file define a `REQUIRED_INTEGRATION=...` variable, so we know
    which data integrations to skip.
    """
    test_file_name = os.path.splitext(os.path.basename(request.fspath))[
        0
    ]  # The extension is stripped out.

    assert test_file_name in allowed_data_integrations_by_file, (
        "%s.py has not specified what data integrations it's allowed to run with, please add those "
        "to the dict in `data_integration_tests/conftest.py`" % test_file_name
    )

    allowed_data_integrations = allowed_data_integrations_by_file[test_file_name]
    if data_integration._metadata.service not in allowed_data_integrations:
        pytest.skip(
            "Skipped for data integration `%s`, since it is not of type `%s`."
            % (data_integration._metadata.name, ",".join(allowed_data_integrations))
        )