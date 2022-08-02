package airflow

import (
	"context"
	"fmt"
	"time"

	operator_db "github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	pollCompileAirflowInterval = 500 * time.Millisecond
	pollCompileAirflowTimeout  = 30 * time.Second
)

// ScheduleWorkflow schedules a workflow with an Airflow backend.
// It returns the Airflow DAG Python spec file and an error, if any.
//
// Assumptions:
//   - All of the in-memory fields of `dag` are set.
//   - All of the in-memory fields of each Operator in `dag.Operators` are set.
func ScheduleWorkflow(
	ctx context.Context,
	dag *workflow_dag.DBWorkflowDag,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	vault vault.Vault,
	db database.Database,
	workflowDagWriter workflow_dag.Writer,
) ([]byte, error) {
	// Generate an Airflow DAG ID
	dagId := generateDagId(dag.Metadata.Name)

	// Prepare the storage credentials, so it can be accessed from Airflow
	airflowStorageConfig, err := prepareStorageConfig(ctx, dag, storageConfig, vault)
	if err != nil {
		return nil, err
	}

	operatorIDs := make([]uuid.UUID, 0, len(dag.Operators))
	for id := range dag.Operators {
		operatorIDs = append(operatorIDs, id)
	}

	artifactIDs := make([]uuid.UUID, 0, len(dag.Artifacts))
	for id := range dag.Artifacts {
		artifactIDs = append(artifactIDs, id)
	}

	// Generate storage path prefixes for operator metadata, artifact content,
	// and artifact metadata. At runtime, the Airflow DAG run ID is appended to
	// the relevant path prefix to form a unique storage path without requiring
	// coordination between Airflow and the Aqueduct server.
	operatorToMetadataPathPrefix := generateStoragePathPrefixes(operatorIDs)
	artifactToContentPathPrefix := generateStoragePathPrefixes(artifactIDs)
	artifactToMetadataPathPrefix := generateStoragePathPrefixes(artifactIDs)

	operatorToTask := make(map[uuid.UUID]string, len(dag.Operators))
	taskToJobSpec := make(map[string]job.Spec, len(dag.Operators))

	// Generate job spec for each Airflow task
	for _, op := range dag.Operators {
		inputArtifacts := make([]artifact.Artifact, 0, len(op.Inputs))
		inputContentPathPrefixes := make([]string, 0, len(op.Inputs))
		inputMetadataPathPrefixes := make([]string, 0, len(op.Inputs))
		for _, artifactId := range op.Inputs {
			dbInputArtifact, ok := dag.Artifacts[artifactId]
			if !ok {
				return nil, errors.Newf("cannot find artifact with ID %v", artifactId)
			}

			contentPath := artifactToContentPathPrefix[artifactId]
			metadataPath := artifactToMetadataPathPrefix[artifactId]

			inputArtifact, err := artifact.NewArtifact(
				dbInputArtifact,
				contentPath,
				metadataPath,
				nil,
				storageConfig,
				nil,
			)
			if err != nil {
				return nil, err
			}

			inputArtifacts = append(inputArtifacts, inputArtifact)
			inputContentPathPrefixes = append(inputContentPathPrefixes, contentPath)
			inputMetadataPathPrefixes = append(inputMetadataPathPrefixes, metadataPath)
		}

		outputArtifacts := make([]artifact.Artifact, 0, len(op.Outputs))
		outputContentPathPrefixes := make([]string, 0, len(op.Outputs))
		outputMetadataPathPrefixes := make([]string, 0, len(op.Outputs))
		for _, artifactId := range op.Outputs {
			dbOutputArtifact, ok := dag.Artifacts[artifactId]
			if !ok {
				return nil, errors.Newf("cannot find artifact with ID %v", artifactId)
			}

			contentPath := artifactToContentPathPrefix[artifactId]
			metadataPath := artifactToMetadataPathPrefix[artifactId]

			outputArtifact, err := artifact.NewArtifact(
				dbOutputArtifact,
				contentPath,
				metadataPath,
				nil,
				storageConfig,
				nil,
			)
			if err != nil {
				return nil, err
			}

			outputArtifacts = append(outputArtifacts, outputArtifact)
			outputContentPathPrefixes = append(outputContentPathPrefixes, contentPath)
			outputMetadataPathPrefixes = append(outputMetadataPathPrefixes, metadataPath)
		}

		airflowOperator, err := operator.NewOperator(
			ctx,
			op,
			inputArtifacts,
			inputContentPathPrefixes,
			inputMetadataPathPrefixes,
			outputArtifacts,
			outputContentPathPrefixes,
			outputMetadataPathPrefixes,
			nil,
			jobManager,
			vault,
			&airflowStorageConfig,
			db,
		)
		if err != nil {
			return nil, err
		}

		jobSpec := airflowOperator.JobSpec()

		taskId := generateTaskId(airflowOperator.Name())

		operatorToTask[airflowOperator.ID()] = taskId
		taskToJobSpec[taskId] = jobSpec
	}

	taskEdges, err := computeEdges(dag.Operators, operatorToTask)
	if err != nil {
		return nil, err
	}

	operatorMetadataPath := fmt.Sprintf("compile-airflow-metadata-%s", uuid.New().String())
	operatorOutputPath := fmt.Sprintf("compile-airflow-output-%s", uuid.New().String())

	defer func() {
		go utils.CleanupStorageFiles(ctx, storageConfig, []string{operatorMetadataPath, operatorOutputPath})
	}()

	jobName := fmt.Sprintf("compile-airflow-operator-%s", uuid.New().String())
	jobSpec, err := job.NewCompileAirflowSpec(
		jobName,
		storageConfig,
		operatorMetadataPath,
		operatorOutputPath,
		dagId,
		string(dag.Metadata.Schedule.CronSchedule),
		taskToJobSpec,
		taskEdges,
	)
	if err != nil {
		return nil, err
	}

	if err := jobManager.Launch(ctx, jobSpec.JobName(), jobSpec); err != nil {
		return nil, err
	}

	jobStatus, err := job.PollJob(ctx, jobSpec.JobName(), jobManager, pollCompileAirflowInterval, pollCompileAirflowTimeout)
	if err != nil {
		return nil, err
	}

	if jobStatus == shared.FailedExecutionStatus {
		return nil, errors.New("Compile Airflow job failed.")
	}

	var execState shared.ExecutionState
	if err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		operatorMetadataPath,
		&execState,
	); err != nil {
		return nil, err
	}

	if execState.Status == shared.FailedExecutionStatus {
		return nil, errors.Newf("Compile Airflow job failed: %v \n logs: %v \n failure type: %v", execState.Error, execState.UserLogs, execState.FailureType)
	}

	airflowDagFile, err := storage.NewStorage(storageConfig).Get(ctx, operatorOutputPath)
	if err != nil {
		return nil, err
	}

	// Update the AirflowRuntimeConfig for `dag`
	newRuntimeConfig := dag.EngineConfig
	newRuntimeConfig.AirflowConfig.OperatorToTask = operatorToTask
	newRuntimeConfig.AirflowConfig.OperatorMetadataPathPrefix = operatorToMetadataPathPrefix
	newRuntimeConfig.AirflowConfig.ArtifactContentPathPrefix = artifactToContentPathPrefix
	newRuntimeConfig.AirflowConfig.ArtifactMetadataPathPrefix = artifactToMetadataPathPrefix

	_, err = workflowDagWriter.UpdateWorkflowDag(
		ctx,
		dag.Id,
		map[string]interface{}{
			workflow_dag.EngineConfigColumn: &newRuntimeConfig,
		},
		db,
	)
	if err != nil {
		return nil, err
	}

	return airflowDagFile, nil
}

// prepareStorageConfig returns the StorageConfig so operators can access storage from the
// Airflow engine.
func prepareStorageConfig(
	ctx context.Context,
	dag *workflow_dag.DBWorkflowDag,
	storageConfig *shared.StorageConfig,
	vault vault.Vault,
) (shared.StorageConfig, error) {
	emptyStorageConf := shared.StorageConfig{}

	authConf, err := auth.ReadConfigFromSecret(ctx, dag.EngineConfig.AirflowConfig.IntegrationId, vault)
	if err != nil {
		return emptyStorageConf, err
	}

	airflowConf, err := parseConfig(authConf)
	if err != nil {
		return emptyStorageConf, err
	}

	if storageConfig.Type != shared.S3StorageType {
		return emptyStorageConf, errors.New("The StorageType must be S3 to use the Airflow engine.")
	}

	return shared.StorageConfig{
		Type: storageConfig.Type,
		S3Config: &shared.S3Config{
			Region:             storageConfig.S3Config.Region,
			Bucket:             storageConfig.S3Config.Bucket,
			CredentialsPath:    airflowConf.S3CredentialsPath,
			CredentialsProfile: airflowConf.S3CredentialsProfile,
		},
	}, nil
}

// generateStoragePathPrefixes generates a storage path prefix for each ID.
func generateStoragePathPrefixes(ids []uuid.UUID) map[uuid.UUID]string {
	paths := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		paths[id] = uuid.NewString()
	}
	return paths
}

func computeEdges(operators map[uuid.UUID]operator_db.DBOperator, operatorToTask map[uuid.UUID]string) (map[string][]string, error) {
	artifactToSrc := map[uuid.UUID]string{}
	artifactToDests := map[uuid.UUID][]string{}

	for _, op := range operators {
		taskId, ok := operatorToTask[op.Id]
		if !ok {
			return nil, errors.Newf("Unable to find task ID for operator %v", op.Id)
		}

		for _, outputArtifact := range op.Outputs {
			artifactToSrc[outputArtifact] = taskId
		}

		for _, inputArtifact := range op.Inputs {
			artifactToDests[inputArtifact] = append(artifactToDests[inputArtifact], taskId)
		}
	}

	taskEdges := map[string][]string{}

	for artifactId, srcTask := range artifactToSrc {
		destTasks, ok := artifactToDests[artifactId]
		if !ok {
			// This artifact has no downstream operators
			continue
		}

		// There is an implicit edge between `srcTask` and each task in `destTasks` via
		// the artifact `artifactId`.
		taskEdges[srcTask] = append(taskEdges[srcTask], destTasks...)
	}

	return taskEdges, nil
}
