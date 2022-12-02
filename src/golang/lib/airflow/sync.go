package airflow

import (
	"context"
	"time"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// SyncWorkflowDags syncs all dags in `workflowDagIds` with any new
// Airflow dag runs since the last sync. It returns an error, if any.
func SyncWorkflowDags(
	ctx context.Context,
	workflowDagIds []uuid.UUID,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	dagResultRepo repos.DAGResult,
	workflowDagWriter workflow_dag.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	vault vault.Vault,
	DB database.Database,
) error {
	// Read each workflow dag from the database that needs to be synced
	dbDags := make([]workflow_dag.DBWorkflowDag, 0, len(workflowDagIds))
	for _, workflowDagId := range workflowDagIds {
		dbDag, err := utils.ReadWorkflowDagFromDatabase(
			ctx,
			workflowDagId,
			workflowReader,
			workflowDagReader,
			operatorReader,
			artifactReader,
			workflowDagEdgeReader,
			DB,
		)
		if err != nil {
			return err
		}

		dbDags = append(dbDags, *dbDag)
	}

	for _, dbDag := range dbDags {
		if err := syncWorkflowDag(
			ctx,
			&dbDag,
			dagResultRepo,
			workflowDagWriter,
			operatorResultWriter,
			artifactResultWriter,
			vault,
			DB,
		); err != nil {
			log.Errorf("Unable to sync with Airflow for WorkflowDag %v: %v", dbDag.Id, err)
		}
	}

	return nil
}

// syncWorkflowDag fetches the latest runs from Airflow for the workflow dag
// specified and populates the database with the results.
// It returns an error, if any.
func syncWorkflowDag(
	ctx context.Context,
	dbDag *workflow_dag.DBWorkflowDag,
	dagResultRepo repos.DAGResult,
	workflowDagWriter workflow_dag.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	vault vault.Vault,
	DB database.Database,
) error {
	// Read Airflow credentials from vault
	authConf, err := auth.ReadConfigFromSecret(
		ctx,
		dbDag.EngineConfig.AirflowConfig.IntegrationId,
		vault,
	)
	if err != nil {
		return err
	}

	// Create Airflow API client
	cli, err := newClient(ctx, authConf)
	if err != nil {
		return err
	}

	dagsMatch, err := checkForDAGMatch(
		ctx,
		cli,
		dbDag,
		workflowDagWriter,
		DB,
	)
	if err != nil {
		return err
	}

	if !dagsMatch {
		// Skip syncing if the dags do not match
		return errors.New("The Airflow DAG does not match the Aqueduct DAG, so the workflow dag cannot be synced.")
	}

	// Get all Airflow DAG runs for `dag`
	dagRuns, err := cli.getDagRuns(dbDag.EngineConfig.AirflowConfig.DagId)
	if err != nil {
		return err
	}

	dagResults, err := dagResultRepo.GetByWorkflow(ctx, dbDag.WorkflowId, DB)
	if err != nil {
		return err
	}

	dagCreatedAtTimes := make([]time.Time, 0, len(dagResults))
	for _, dagResult := range dagResults {
		dagCreatedAtTimes = append(dagCreatedAtTimes, dagResult.CreatedAt)
	}

	for _, dagRun := range dagRuns {
		// TODO: What if this dagRun corresponds to a previous WorkflowDag?

		// Check if this DagRun has already been synced.
		// We reasonably assume that no 2 Airflow DagRuns have the same start date, because
		// the DagRun start date is measured in nanoseconds.
		if ok := timeInSlice(dagRun.GetStartDate(), dagCreatedAtTimes); ok {
			// A DagRun with the same start time has already been registered, so skip this
			continue
		}

		if *dagRun.State != airflow.DAGSTATE_SUCCESS &&
			*dagRun.State != airflow.DAGSTATE_FAILED {
			// DagRun is in either DAGSTATE_QUEUED or DAGSTATE_RUNNING,
			// i.e. it has not finished yet, so skip it.
			continue
		}

		// Populate database with WorkflowDagResult for this DagRun
		if err := syncWorkflowDagResult(
			ctx,
			cli,
			dbDag,
			&dagRun,
			dagResultRepo,
			operatorResultWriter,
			artifactResultWriter,
			DB,
		); err != nil {
			return err
		}
	}

	return nil
}

// syncWorkflowDagResult populates the database with a WorkflowDagResult and related
// OperatorResult(s) and ArtifactResult(s) for the Airflow DagRun `run` of the DBWorkflowDag `dbDag`.
// It returns an error, if any.
func syncWorkflowDagResult(
	ctx context.Context,
	cli *client,
	dbDag *workflow_dag.DBWorkflowDag,
	run *airflow.DAGRun,
	dagResultRepo repos.DAGResult,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	DB database.Database,
) error {
	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	dagResult, err := createDAGResult(
		ctx,
		dbDag,
		run,
		dagResultRepo,
		txn,
	)
	if err != nil {
		return err
	}

	// Get Airflow task states
	taskToState, err := cli.getTaskStates(run.GetDagId(), run.GetDagRunId())
	if err != nil {
		return err
	}

	for _, op := range dbDag.Operators {
		// Map Airflow task state to operator execution status
		taskID, ok := dbDag.EngineConfig.AirflowConfig.OperatorToTask[op.Id]
		if !ok {
			return errors.Newf("Unable to determine Airflow task ID for operator %v", op.Id)
		}

		taskState, ok := taskToState[taskID]
		if !ok {
			return errors.Newf("Unable to find Airflow task state for task %s", taskID)
		}

		execStatus := mapTaskStateToStatus(taskState)

		if err := createOperatorResult(
			ctx,
			run.GetDagRunId(),
			dbDag,
			&op,
			execStatus,
			dagResult.ID,
			operatorResultWriter,
			artifactResultWriter,
			txn,
		); err != nil {
			return err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// checkForDAGMatch checks if the Aqueduct workflow DAG `dbDag`
// matches the DAG currently registered with Airflow. They may not match if the user
// updated the workflow and has not yet copied over the updated Airflow DAG file to
// their Airflow server. If the DAGs match, it also updates `dbDag`'s engine config
// in the database.
// It returns a bool whether the DAGs match and an error, if any.
func checkForDAGMatch(
	ctx context.Context,
	cli *client,
	dbDag *workflow_dag.DBWorkflowDag,
	workflowDagWriter workflow_dag.Writer,
	db database.Database,
) (bool, error) {
	if dbDag.EngineConfig.AirflowConfig.MatchesAirflow {
		// We previously confirmed that the DAGs match
		return true, nil
	}

	airflowDag, err := cli.getDag(dbDag.EngineConfig.AirflowConfig.DagId)
	if err != nil {
		return false, err
	}

	// The way we check if the DAGs match is if `dbDag.Id` is one of tags
	// for `airflowDag`, since the workflow dag ID is set as a tag each time
	// the Airflow DAG file is generated.
	for _, tag := range airflowDag.Tags {
		if tag.GetName() == dbDag.Id.String() {
			// The DAGs match so the engine config needs to be updated
			dbDag.EngineConfig.AirflowConfig.MatchesAirflow = true
			_, err = workflowDagWriter.UpdateWorkflowDag(
				ctx,
				dbDag.Id,
				map[string]interface{}{
					workflow_dag.EngineConfigColumn: &dbDag.EngineConfig,
				},
				db,
			)
			if err != nil {
				return true, err
			}

			return true, nil
		}
	}

	return false, nil
}

// timeInSlice returns whether `t` is equal to any of the elements in `s`
func timeInSlice(t time.Time, s []time.Time) bool {
	for _, tt := range s {
		if t.Equal(tt) {
			return true
		}
	}
	return false
}
