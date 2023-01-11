import pandas as pd
import pytest
from aqueduct.constants.enums import CheckSeverity
from aqueduct.error import AqueductError, ArtifactNotFoundException, InvalidUserActionException

from aqueduct import check

from ..shared.data_objects import DataObject
from ..shared.utils import extract, publish_flow_test
from .test_functions.simple.model import dummy_sentiment_model
from .test_metrics.constant.model import constant_metric


@check()
def success_on_single_table_input(df):
    if not isinstance(df, pd.DataFrame):
        raise Exception("Expected dataframe as input to check, got %s" % type(df).__name__)
    return True


@check()
def success_on_single_metric_input(metric):
    if not isinstance(metric, float):
        raise Exception("Expected float as input to check, got %s" % type(metric).__name__)
    return True


@check()
def success_on_multiple_mixed_inputs(metric, df):
    if not isinstance(metric, float):
        raise Exception("Expected float as input to check, got %s" % type(metric).__name__)
    if not isinstance(df, pd.DataFrame):
        raise Exception("Expected dataframe as input to check, got %s" % type(df).__name__)
    return True


def test_check_on_table(client, flow_name, data_integration, engine):
    """Test check on a function operator."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    check_artifact = success_on_single_table_input(table_artifact)
    assert check_artifact.get()

    publish_flow_test(
        client,
        check_artifact,
        name=flow_name(),
        engine=engine,
    )


def test_check_on_metric(client, flow_name, data_integration, engine):
    """Test check on a metric operator."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    metric = constant_metric(table_artifact)

    check_artifact = success_on_single_metric_input(metric)
    assert check_artifact.get()

    publish_flow_test(
        client,
        check_artifact,
        name=flow_name(),
        engine=engine,
    )


def test_check_on_multiple_mixed_inputs(client, flow_name, data_integration, engine):
    """Test check on multiple tables and metrics."""
    table_artifact1 = extract(data_integration, DataObject.SENTIMENT)
    metric = constant_metric(table_artifact1)

    table_artifact2 = extract(data_integration, DataObject.SENTIMENT)
    table = dummy_sentiment_model(table_artifact2)

    check_artifact = success_on_multiple_mixed_inputs(metric, table)
    assert check_artifact.get()

    publish_flow_test(
        client,
        check_artifact,
        name=flow_name(),
        engine=engine,
    )


def test_edit_check(client, data_integration):
    """Test that checks can be edited by replacing with the same name."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @check()
    def check_op(df):
        return False

    failed_check = check_op(table_artifact)
    assert not failed_check.get()

    @check()
    def check_op(df):
        return True

    success_check = check_op(table_artifact)
    assert success_check.get()

    # Attempting to fetch the previous check artifact should fail, since its been overwritten!
    with pytest.raises(ArtifactNotFoundException):
        failed_check.get()


def test_delete_check(client, data_integration):
    """Test that checks can be deleted by name."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    with pytest.raises(InvalidUserActionException):
        table_artifact.remove_check(name="nonexistant_check")

    check_artifact_on_sql = success_on_single_table_input(table_artifact)
    table_artifact.remove_check(name="success_on_single_table_input")
    with pytest.raises(ArtifactNotFoundException):
        check_artifact_on_sql.get()

    metric_artifact = constant_metric(table_artifact)
    check_artifact_on_metric = success_on_single_metric_input(metric_artifact)
    metric_artifact.remove_check(name="success_on_single_metric_input")
    with pytest.raises(ArtifactNotFoundException):
        check_artifact_on_metric.get()


def test_check_wrong_input_type(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    # User function receives a dataframe when it's expecting a metric.
    with pytest.raises(AqueductError):
        check_artifact = success_on_single_metric_input(table_artifact)

    # TODO(ENG-862): the following code should not surface an internal error,
    #  since its the user's fault.
    # Running a function operator on a check output, which is not allowed.
    check_artifact = success_on_single_table_input(table_artifact)
    with pytest.raises(Exception):
        dummy_sentiment_model(check_artifact)


def test_check_wrong_number_of_inputs(client, data_integration):
    table_artifact1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact2 = extract(data_integration, DataObject.SENTIMENT)

    # TODO(ENG-863): Do we want a more specific error here?
    with pytest.raises(AqueductError):
        success_on_single_table_input(table_artifact1, table_artifact2)


def test_check_with_numpy_bool_output(client, data_integration):
    table_artifact = extract(data_integration, DataObject.CHURN)

    @check()
    def success_check_return_numpy_bool(df):
        return df["total_charges"].mean() < 2500

    check_artifact = success_check_return_numpy_bool(table_artifact)
    assert check_artifact.get()


def test_check_with_series_output(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @check()
    def success_check_return_series_of_booleans(df):
        return pd.Series([True, True, True])

    @check()
    def failure_check_return_series_of_booleans(df):
        return pd.Series([True, False, True])

    passed = success_check_return_series_of_booleans(table_artifact)
    assert passed.get()

    failed = failure_check_return_series_of_booleans(table_artifact)
    assert not failed.get()

    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[table_artifact, passed, failed],
        engine=engine,
    )


def test_check_failure_with_varying_severity(client, flow_name, data_integration, engine):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    # An error check will fail the workflow, but a warning check will not.
    @check(severity=CheckSeverity.WARNING)
    def failure_nonblocking_check(df):
        return False

    @check(severity=CheckSeverity.ERROR)
    def failure_blocking_check(df):
        return False

    nonblocking_check = failure_nonblocking_check(table_artifact)

    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[table_artifact, nonblocking_check],
        engine=engine,
    )

    # In eager execution, this check should fail before we can publish the flow.
    with pytest.raises(AqueductError):
        failure_blocking_check(table_artifact)