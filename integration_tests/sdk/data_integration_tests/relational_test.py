from typing import List

import pandas as pd
import pytest
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.generic_artifact import GenericArtifact
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import AqueductError, InvalidUserActionException, InvalidUserArgumentException
from aqueduct.models.operators import RelationalDBExtractParams

from aqueduct import LoadUpdateMode, metric, op

from ..shared.demo_db import demo_db_tables
from ..shared.relational import SHORT_SENTIMENT_SQL_QUERY
from ..shared.utils import generate_table_name, publish_flow_test
from .save import save


def _create_successful_sql_artifacts(
    client,
    data_integration,
    wrap_query_in_extract_params_struct: bool = False,
) -> List[BaseArtifact]:
    """Tests and returns artifacts for two types of sql queries: basic and chained.

    Every artifact is saved to a random table with update_mode='replace'.
    """
    hotel_reviews_query = "SELECT * FROM hotel_reviews"
    chained_query = [
        "SELECT * FROM hotel_reviews",
        "SELECT review, review_date FROM $ WHERE reviewer_nationality ='{{nationality}}'",
        "SELECT review FROM $",
    ]

    if wrap_query_in_extract_params_struct:
        hotel_reviews_query = RelationalDBExtractParams(query=hotel_reviews_query)
        chained_query = RelationalDBExtractParams(queries=chained_query)

    # Test a successful basic sql query.
    hotel_reviews_table = data_integration.sql(hotel_reviews_query)
    assert list(hotel_reviews_table.get()) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]
    assert hotel_reviews_table.get().shape[0] == 100

    # Test a successful chain query.
    client.create_param("nationality", default=" United Kingdom ")
    chained_query_result = data_integration.sql(chained_query)
    expected_chained_query_result = data_integration.sql(
        "SELECT review FROM hotel_reviews WHERE reviewer_nationality=' United Kingdom '",
    )
    assert expected_chained_query_result.get().equals(chained_query_result.get())

    artifacts = [hotel_reviews_table, chained_query_result]
    for artifact in artifacts:
        save(data_integration, artifact, generate_table_name(), LoadUpdateMode.REPLACE)
    return artifacts


def _publish_saves_and_check(client, flow_name, validator, artifacts: List[BaseArtifact]):
    flow = publish_flow_test(
        client,
        artifacts,
        name=flow_name(),
        engine=None,
    )
    for artifact in artifacts:
        validator.check_saved_artifact_data(flow, artifact.id(), expected_data=artifact.get())


def test_sql_integration_query_and_save(client, flow_name, data_integration, validator):
    artifacts = _create_successful_sql_artifacts(client, data_integration)
    _publish_saves_and_check(client, flow_name, validator, artifacts)


def test_sql_integration_query_and_save_relationaldbextractparams(
    client, flow_name, data_integration, validator
):
    artifacts = _create_successful_sql_artifacts(
        client, data_integration, wrap_query_in_extract_params_struct=True
    )
    _publish_saves_and_check(client, flow_name, validator, artifacts)


def test_sql_integration_artifact_with_custom_metadata(
    client, flow_name, data_integration, validator
):
    # TODO: validate custom descriptions once we can fetch descriptions easily.
    artifact = data_integration.sql(
        "SELECT * FROM hotel_reviews", name="Test Artifact", description="This is a description"
    )
    assert artifact.name() == "Test Artifact artifact"

    flow = publish_flow_test(client, artifact, name=flow_name(), engine=None)
    validator.check_artifact_was_computed(flow, "Test Artifact artifact")


def test_sql_integration_failed_query(client, data_integration):
    # Sql query is malformed.
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_integration.sql("SELECT * FROM ")

    # SQL error happens at execution time (table missing).
    with pytest.raises(AqueductError, match="Preview Execution Failed"):
        data_integration.sql("SELECT * FROM missing_table")


def test_sql_integration_table_retrieval(client, data_integration):
    df = data_integration.table(name="hotel_reviews")
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_sql_integration_list_tables(client, data_integration):
    tables = data_integration.list_tables()

    for expected_table in demo_db_tables():
        assert tables["tablename"].str.contains(expected_table, case=False).sum() > 0


def test_sql_today_tag(client, data_integration):
    table_artifact_today = data_integration.sql(
        query="select * from hotel_reviews where review_date = {{today}}"
    )
    assert table_artifact_today.get().empty
    table_artifact_not_today = data_integration.sql(
        query="select * from hotel_reviews where review_date < {{today}}"
    )
    assert len(table_artifact_not_today.get()) == 100


def test_sql_query_with_parameter(client, data_integration):
    # Missing parameters.
    with pytest.raises(InvalidUserArgumentException):
        _ = data_integration.sql(query="select * from {{missing_parameter}}")

    # The parameter is not a string type.
    _ = client.create_param("table_name", default=1234)
    with pytest.raises(InvalidUserArgumentException):
        _ = data_integration.sql(query="select * from {{ table_name }}")

    client.create_param("table_name", default="hotel_reviews")
    table_artifact = data_integration.sql(query="select * from {{ table_name }}")

    expected_table_artifact = data_integration.sql(query="select * from hotel_reviews")
    assert table_artifact.get().equals(expected_table_artifact.get())
    expected_table_artifact = data_integration.sql(query="select * from customer_activity")
    assert table_artifact.get(parameters={"table_name": "customer_activity"}).equals(
        expected_table_artifact.get()
    )

    # Trigger the parameter with invalid values.
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"table_name": ["this is the incorrect type"]})
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"non-existant parameter": "blah"})


def test_sql_query_with_multiple_parameters(client, flow_name, data_integration, engine):
    _ = client.create_param("table_name", default="hotel_reviews")
    nationality = client.create_param(
        "reviewer-nationality", default="United Kingdom"
    )  # check that dashes work.
    table_artifact = data_integration.sql(
        query="select * from {{ table_name }} where reviewer_nationality='{{ reviewer-nationality }}' and review_date < {{ today}}"
    )
    expected_table_artifact = data_integration.sql(
        "select * from hotel_reviews where reviewer_nationality='United Kingdom' and review_date < {{today}}"
    )
    assert table_artifact.get().equals(expected_table_artifact.get())
    expected_table_artifact = data_integration.sql(
        "select * from hotel_reviews where reviewer_nationality='Australia' and review_date < {{today}}"
    )
    assert table_artifact.get(parameters={"reviewer-nationality": "Australia"}).equals(
        expected_table_artifact.get()
    )

    # Use the parameters in another operator.
    @metric
    def noop(sql_output, param):
        return len(param)

    result = noop(table_artifact, nationality)
    assert result.get() == len(nationality.get())
    assert result.get(parameters={"reviewer-nationality": "Australia"}) == len("Australia")

    publish_flow_test(client, name=flow_name(), artifacts=[result], engine=engine)


def test_sql_query_user_vs_builtin_precedence(client, data_integration):
    """If a user defines an expansion that collides with a built-in one, the user-defined one should take precedence."""
    table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > {{today}}"
    )
    builtin_result = table_artifact.get()

    datestring = "'2016-01-01'"
    _ = client.create_param("today", datestring)
    table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > {{today}}"
    )
    user_param_result = table_artifact.get()
    assert not builtin_result.equals(user_param_result)

    expected_table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > %s" % datestring
    )
    assert user_param_result.equals(expected_table_artifact.get())


def test_sql_integration_save_wrong_data_type(client, flow_name, data_integration):
    # Try to save a numeric artifact.
    num_param = client.create_param("number", default=123)
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to save non-relational data into relational data store",
    ):
        save(data_integration, num_param, generate_table_name(), LoadUpdateMode.REPLACE)

    # Save a generic artifact that is actually a string. This won't fail at save() time,
    # but instead when the flow is published.
    @op
    def foo():
        return "asdf"

    string_artifact = foo.lazy()
    assert isinstance(string_artifact, GenericArtifact)
    save(data_integration, string_artifact, generate_table_name(), LoadUpdateMode.REPLACE)
    publish_flow_test(
        client,
        string_artifact,
        name=flow_name(),
        engine=None,
        expected_statuses=ExecutionStatus.FAILED,
    )


def test_sql_integration_save_with_different_update_modes(
    client, flow_name, data_integration, engine, validator
):
    table_1_save_name = generate_table_name()
    table_2_save_name = generate_table_name()

    table = data_integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    extracted_table_data = table.get()
    save(data_integration, table, table_1_save_name, LoadUpdateMode.REPLACE)

    # This will create the table.
    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact_data(flow, table.id(), expected_data=extracted_table_data)

    # Change to append mode.
    save(data_integration, table, table_1_save_name, LoadUpdateMode.APPEND)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    # Redundant append mode change
    save(data_integration, table, table_1_save_name, LoadUpdateMode.APPEND)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=pd.concat(
            [extracted_table_data, extracted_table_data, extracted_table_data], ignore_index=True
        ),
    )

    # Create a different table from the same artifact.
    save(data_integration, table, table_2_save_name, LoadUpdateMode.REPLACE)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact_data(
        flow,
        table.id(),
        expected_data=extracted_table_data,
    )

    validator.check_saved_update_mode_changes(
        flow,
        expected_updates=[
            (table_2_save_name, LoadUpdateMode.REPLACE),
            (table_1_save_name, LoadUpdateMode.APPEND),
            (table_1_save_name, LoadUpdateMode.REPLACE),
        ],
    )