package operator

import (
	"context"
	"fmt"

	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/check"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type baseOperator struct {
	dbOperator *operator.DBOperator

	// These fields are set to nil in the preview case.
	resultWriter operator_result.Writer
	resultID     uuid.UUID

	metadataPath string
	jobName      string

	inputs          []artifact.Artifact
	outputs         []artifact.Artifact
	inputExecPaths  []*utils.ExecPaths
	outputExecPaths []*utils.ExecPaths

	// The operator is cache-aware if this is non-nil.
	previewArtifactCacheManager artifact.PreviewCacheManager
	jobManager                  job.JobManager
	vaultObject                 vault.Vault
	storageConfig               *shared.StorageConfig
	db                          database.Database

	// This cannot be set if the operator is cache-aware, since this only happens in non-preview paths.
	resultsPersisted bool
	isPreview        bool
}

func (bo *baseOperator) Type() operator.Type {
	return bo.dbOperator.Spec.Type()
}

func (bo *baseOperator) Name() string {
	return bo.dbOperator.Name
}

func (bo *baseOperator) ID() uuid.UUID {
	return bo.dbOperator.Id
}

func (bo *baseOperator) MetadataPath() string {
	return bo.metadataPath
}

// A catch-all for execution states that are the system's fault.
// Logs an internal message so that we can debug.
func unknownSystemFailureExecState(err error, logMsg string) *shared.ExecutionState {
	log.Errorf("Execution had system failure: %s. %v", logMsg, err)

	failureType := shared.SystemFailure
	return &shared.ExecutionState{
		Status:      shared.FailedExecutionStatus,
		FailureType: &failureType,
		Error: &shared.Error{
			Context: fmt.Sprintf("%v", err),
			Tip:     shared.TipUnknownInternalError,
		},
	}
}

// For each output artifact, copy over the contents of the content and metadata paths.
// This should only ever be used for preview routes. Returns whether this operator has succeeded.
// If it does not, the operator will fall back on traditional execution, overwriting anything we did here.
func (bo *baseOperator) executeUsingCachedResult(ctx context.Context, cachedResultByLogicalID map[uuid.UUID]artifact.PreviewCacheEntry) bool {
	// Assumption: there is only one output artifact.
	for i, outputArtifact := range bo.outputs {
		cachedResult := cachedResultByLogicalID[outputArtifact.LogicalID()]
		err := utils.CopyPathInStorage(ctx, bo.storageConfig, cachedResult.ArtifactContentPath, bo.outputExecPaths[i].ArtifactContentPath)
		if err != nil {
			return false
		}

		err = utils.CopyPathInStorage(ctx, bo.storageConfig, cachedResult.ArtifactMetadataPath, bo.outputExecPaths[i].ArtifactMetadataPath)
		if err != nil {
			return false
		}

		err = utils.CopyPathInStorage(ctx, bo.storageConfig, cachedResult.OpMetadataPath, bo.outputExecPaths[i].OpMetadataPath)
		if err != nil {
			return false
		}
	}
	return true
}

func (bo *baseOperator) launch(ctx context.Context, spec job.Spec) error {
	// Check if this operator can use previously cached results instead of computing for scratch.
	if bo.previewArtifactCacheManager != nil {
		outputArtifactLogicalIDs := make([]uuid.UUID, 0, len(bo.outputs))
		for _, outputArtifact := range bo.outputs {
			outputArtifactLogicalIDs = append(outputArtifactLogicalIDs, outputArtifact.LogicalID())
		}

		allCached, cachedResultByID, err := bo.previewArtifactCacheManager.GetMulti(ctx, outputArtifactLogicalIDs)
		if err != nil {
			log.Errorf("Unable to fetch output artifact ids %v. Error: %v", outputArtifactLogicalIDs, err)
		}

		// Only use cached results immediately if all output artifacts are cached.
		if allCached {
			succeeded := bo.executeUsingCachedResult(ctx, cachedResultByID)
			if succeeded {
				return nil
			}
		}
	}

	return bo.jobManager.Launch(ctx, spec.JobName(), spec)
}

// fetchExecState assumes that the operator has been computed already.
func (bo *baseOperator) fetchExecState(ctx context.Context) *shared.ExecutionState {
	var execState shared.ExecutionState
	err := utils.ReadFromStorage(
		ctx,
		bo.storageConfig,
		bo.metadataPath,
		&execState,
	)
	if err != nil {
		// Treat this as a system internal error since operator metadata was not found
		return unknownSystemFailureExecState(
			err,
			"Unable to read operator metadata from storage. Operator may have failed before writing metadata.",
		)
	}
	return &execState
}

// GetExecState takes a more wholelistic view of the operator's status than the job manager does,
// and can be called at any time. Because of this, the logic for figuring out the correct state is
// a little more involved.
func (bo *baseOperator) GetExecState(ctx context.Context) (*shared.ExecutionState, error) {
	if bo.jobName == "" {
		return nil, errors.Newf("Internal error: a job name was not set for this operator.")
	}

	status, err := bo.jobManager.Poll(ctx, bo.jobName)
	if err != nil {
		// If the job does not exist, this could mean that
		// 1) it is hasn't been run yet (pending),
		// 2) it has run already at sometime in the past, but has been garbage collected
		// 3) it has run already at sometime in the past, but did not go through the job manager.
		//    (this can happen when the output artifacts have already been cached).
		if err == job.ErrJobNotExist {
			// Check whether the operator has actually completed.
			if utils.ObjectExistsInStorage(ctx, bo.storageConfig, bo.metadataPath) {
				return bo.fetchExecState(ctx), nil
			}

			// Otherwise, this job has not run yet and is in a pending state.
			return &shared.ExecutionState{
				Status: shared.PendingExecutionStatus,
			}, nil
		} else {
			// This is just an internal polling error state.
			return unknownSystemFailureExecState(err, "Unable to poll job manager."), nil
		}
	} else {
		// The job just completed, so we know we can fetch the results (succeeded/failed).
		if status == shared.FailedExecutionStatus || status == shared.SucceededExecutionStatus {
			return bo.fetchExecState(ctx), nil
		}

		// The job must exist at this point, but it hasn't completed (running).
		return &shared.ExecutionState{
			Status: shared.RunningExecutionStatus,
		}, nil
	}
}

func updateOperatorResultAfterComputation(
	ctx context.Context,
	execState *shared.ExecutionState,
	opResultWriter operator_result.Writer,
	opResultID uuid.UUID,
	db database.Database,
) {
	changes := map[string]interface{}{
		operator_result.StatusColumn:    execState.Status,
		operator_result.ExecStateColumn: execState,
	}

	_, err := opResultWriter.UpdateOperatorResult(
		ctx,
		opResultID,
		changes,
		db,
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"changes": changes,
			},
		).Errorf("Unable to update operator result metadata: %v", err)
	}
}

func (bo *baseOperator) InitializeResult(ctx context.Context, dagResultID uuid.UUID) error {
	if bo.resultWriter == nil {
		return errors.New("Operator's result writer cannot be nil.")
	}

	operatorResult, err := bo.resultWriter.CreateOperatorResult(
		ctx,
		dagResultID,
		bo.ID(),
		bo.db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create operator result record.")
	}
	bo.resultID = operatorResult.Id
	return nil
}

func (bo *baseOperator) PersistResult(ctx context.Context) error {
	if bo.isPreview {
		// Don't persist any result for preview operators.
		return errors.Newf("Operator %s cannot be persisted, as it is being previewed.")
	}

	if bo.previewArtifactCacheManager != nil {
		return errors.Newf("Operator %s is cache-aware, so it cannot be persisted.", bo.Name())
	}

	if bo.resultsPersisted {
		return errors.Newf("Operator %s was already persisted!", bo.Name())
	}

	execState, err := bo.GetExecState(ctx)
	if err != nil {
		return err
	}
	if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
		return errors.Newf("Operator %s has neither succeeded or failed, so it does not have results that can be persisted.", bo.Name())
	}

	// Best effort writes after this point.
	updateOperatorResultAfterComputation(
		ctx,
		execState,
		bo.resultWriter,
		bo.resultID,
		bo.db,
	)

	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.PersistResult(ctx, execState)
		if err != nil {
			log.Errorf("Error occurred when persisting artifact %s.", outputArtifact.Name())
		}
	}
	bo.resultsPersisted = true
	return nil
}

func (bo *baseOperator) Finish(ctx context.Context) {
	// Delete the operator's metadata path only if it was already copied into the operator_result's table.
	// Otherwise, the artifact preview cache manager will handle the deletion.
	if bo.resultsPersisted {
		utils.CleanupStorageFile(ctx, bo.storageConfig, bo.metadataPath)
	}

	for _, outputArtifact := range bo.outputs {
		outputArtifact.Finish(ctx)
	}
}

// Any operator that runs a python function serialized from storage should use this instead of baseOperator.
type baseFunctionOperator struct {
	baseOperator
}

func (bfo *baseFunctionOperator) Finish(ctx context.Context) {
	// If the operator ran in preview mode, cleanup the serialized function.
	if bfo.isPreview {
		utils.CleanupStorageFile(ctx, bfo.storageConfig, bfo.dbOperator.Spec.Function().StoragePath)
	}

	bfo.baseOperator.Finish(ctx)
}

const (
	defaultFunctionEntryPointFile   = "model.py"
	defaultFunctionEntryPointClass  = "Function"
	defaultFunctionEntryPointMethod = "predict"
)

func (bfo *baseFunctionOperator) jobSpec(
	fn *function.Function,
	checkSeverity *check.Level,
) job.Spec {
	entryPoint := fn.EntryPoint
	if entryPoint == nil {
		entryPoint = &function.EntryPoint{
			File:      defaultFunctionEntryPointFile,
			ClassName: defaultFunctionEntryPointClass,
			Method:    defaultFunctionEntryPointMethod,
		}
	}

	inputArtifactTypes := make([]db_artifact.Type, 0, len(bfo.inputs))
	outputArtifactTypes := make([]db_artifact.Type, 0, len(bfo.outputs))
	for _, inputArtifact := range bfo.inputs {
		inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Type())
	}
	for _, outputArtifact := range bfo.outputs {
		outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Type())
	}

	inputContentPaths, inputMetadataPaths := unzipExecPathsToRawPaths(bfo.inputExecPaths)
	outputContentPaths, outputMetadataPaths := unzipExecPathsToRawPaths(bfo.outputExecPaths)
	return &job.FunctionSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.FunctionJobType,
			bfo.jobName,
			*bfo.storageConfig,
			bfo.metadataPath,
		),
		FunctionPath: fn.StoragePath,
		/* `FunctionExtractPath` is set by the job manager at launch time. */
		EntryPointFile:      entryPoint.File,
		EntryPointClass:     entryPoint.ClassName,
		EntryPointMethod:    entryPoint.Method,
		CustomArgs:          fn.CustomArgs,
		InputContentPaths:   inputContentPaths,
		InputMetadataPaths:  inputMetadataPaths,
		OutputContentPaths:  outputContentPaths,
		OutputMetadataPaths: outputMetadataPaths,
		InputArtifactTypes:  inputArtifactTypes,
		OutputArtifactTypes: outputArtifactTypes,
		OperatorType:        bfo.Type(),
		CheckSeverity:       checkSeverity,
	}
}
