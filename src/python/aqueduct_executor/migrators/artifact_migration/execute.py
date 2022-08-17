import json
import sys
import cloudpickle as pickle

from aqueduct_executor.operators.artifact_migration.spec import MigrationSpec
from aqueduct_executor.operators.utils.enums import ArtifactType, SerializationType
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.utils import infer_artifact_type


_DEFAULT_ENCODING = "utf8"
_METADATA_ARTIFACT_TYPE_KEY = "artifact_type"
_METADATA_SERIALIZATION_TYPE_KEY = "serialization_type"


def _write_string_output(
    storage: Storage,
    output_path: str,
    output: str,
) -> None:
    storage.put(output_path, output.encode(_DEFAULT_ENCODING))


def _write_pickle_output(
    storage: Storage,
    output_path: str,
    output: Any,
) -> None:
    storage.put(output_path, pickle.dumps(output))


def _write_json_output(
    storage: Storage,
    output_path: str,
    output: Any,
) -> None:
    storage.put(output_path, json.dumps(output).encode(_DEFAULT_ENCODING))


_serialization_function_mapping = {
    SerializationType.JSON: _write_json_output,
    SerializationType.PICKLE: _write_pickle_output,
    SerializationType.STRING: _write_string_output,
}


def run(spec: MigrationSpec) -> None:
    """
    Executes a artifact migration.
    """
    print("Job Spec: \n{}".format(spec.json()))

    storage = parse_storage(spec.storage_config)
    artifact_metadata = {}

    if spec.artifact_type == "table":
        artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY] = ArtifactType.TABLE.value
        artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.TABLE.value
    elif spec.artifact_type == "float":
        artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY] = ArtifactType.NUMERIC.value
        artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value
    elif spec.artifact_type == "boolean":
        artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY] = ArtifactType.BOOL.value
        artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value

        artifact_content = bool(storage.get(spec.content_path))
        storage.put(spec.content_path, json.dumps(artifact_content).encode(_DEFAULT_ENCODING))
    elif spec.artifact_type == "json":
        content = json.loads(storage.get(spec.content_path).decode(_DEFAULT_ENCODING))
        new_artifact_type = infer_artifact_type(content)

        if new_artifact_type == ArtifactType.JSON or new_artifact_type == ArtifactType.STRING:
            artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.STRING.value
        elif new_artifact_type == ArtifactType.BOOL or new_artifact_type == ArtifactType.NUMERIC:
            artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value
        elif new_artifact_type == ArtifactType.PICKLABLE:
            artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.PICKLE.value
        elif new_artifact_type == ArtifactType.DICT or new_artifact_type == ArtifactType.TUPLE:
            try:
                json.dumps(content)
                artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value
            except:
                artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.PICKLE.value
        else:
            raise Exception("Unexpected artifact type %s" % new_artifact_type)

        _serialization_function_mapping[artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY]](
            storage,
            spec.content_path,
            content,
        )

        artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY] = new_artifact_type.value
    
    storage.put(spec.metadata_path, json.dumps(artifact_metadata).encode(_DEFAULT_ENCODING))
