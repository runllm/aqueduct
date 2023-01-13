import pytest
from aqueduct.error import (
    InvalidIntegrationException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from pydantic import ValidationError

from ..shared.data_objects import DataObject
from .extract import extract
from .save import save
from .test_functions.simple.model import dummy_sentiment_model


def test_invalid_source_integration(client):
    with pytest.raises(InvalidIntegrationException):
        client.integration(name="wrong integration name")


def test_invalid_destination_integration(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    with pytest.raises(InvalidIntegrationException):
        data_integration._metadata.name = "bad name"
        save(data_integration, output_artifact)


def test_invalid_connect_integration(client):
    # Name already exists.
    config = {
        "database": "test",
    }
    with pytest.raises(
        InvalidUserActionException, match="An integration with this name already exists."
    ):
        client.connect_integration("aqueduct_demo", "SQLite", config)

    # Service is invalid.
    with pytest.raises(
        InvalidUserArgumentException,
        match="Service argument must match exactly one of the enum values in ServiceType.",
    ):
        client.connect_integration("New Integration", "invalid service", config)

    # Invalid config raises a pydantic error.
    with pytest.raises(ValidationError):
        client.connect_integration("New Integration", "SQLite", {})