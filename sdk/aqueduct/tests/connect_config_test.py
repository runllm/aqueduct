from aqueduct.integrations import S3Config, GCSConfig


def test_consistent_use_as_storage_field_name_for_storage_layer_integrations():
    # Checks that all data connection configs have the same field name for using as storage layer.
    # This is a necessary assumption to enforce for our integration test setup.
    field_name = "use_as_storage"
    assert hasattr(S3Config, field_name)
    assert hasattr(GCSConfig, field_name)
