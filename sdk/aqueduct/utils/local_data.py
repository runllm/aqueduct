import os
from typing import Optional

from aqueduct.constants.enums import ArtifactType, LocalDataTableFormat
from aqueduct.models.local_data import LocalData

from ..error import InvalidUserArgumentException

# The current local data maximum is temporarily set to 1MB.
MAXIMUM_LOCAL_DATA_SIZE = 1000000


def local_data(
    path: str,
    artifact_type: ArtifactType,
    format: Optional[str] = None,
) -> LocalData:
    """Identify the local data which can be passed in as a parameter.

    Args:
        path:
            The path to the data.
        artifact_type:
                The expected type of the data. Currently local_data has support for ArtifactType.TABLE
                and ArtifactType.IMAGE
        format:
                If the artifact type is ArtifactType.TABLE, the user has to specify the table format.
                We currently support "JSON", "CSV", and "Parquet".

    Returns
        A configuration object which you can use to reference data in the local filesystem.
    """

    format_enum = _convert_to_local_data_table_format(format)
    local_data = LocalData(path=path, artifact_type=artifact_type, format=format_enum)
    validate_local_data(local_data)
    return local_data


def validate_local_data(val: LocalData) -> None:
    """Validate LocalData on its file path and types."""
    file_path = val.path
    artifact_type = val.artifact_type
    file_format = val.format

    if not os.path.isfile(file_path):
        raise InvalidUserArgumentException(
            "Given path file '%s' to local data does not exist." % file_path
        )

    if os.path.getsize(file_path) > MAXIMUM_LOCAL_DATA_SIZE:
        raise InvalidUserArgumentException("Currently, the maximum local data size is 1 MB.")
    if artifact_type == ArtifactType.TABLE and file_format is None:
        raise InvalidUserArgumentException(
            "Specify format in order to use local data as TableArtifact."
        )


def _convert_to_local_data_table_format(format: Optional[str]) -> Optional[LocalDataTableFormat]:
    """A simple string -> enum conversion. Returns None if no format provided."""
    if format is None:
        return None

    lowercased_format = format.lower()
    if lowercased_format == LocalDataTableFormat.CSV.value.lower():
        format_enum = LocalDataTableFormat.CSV
    elif lowercased_format == LocalDataTableFormat.JSON.value.lower():
        format_enum = LocalDataTableFormat.JSON
    elif lowercased_format == LocalDataTableFormat.PARQUET.value.lower():
        format_enum = LocalDataTableFormat.PARQUET
    else:
        raise InvalidUserArgumentException("Unsupported local file format `%s`." % format)
    return format_enum
