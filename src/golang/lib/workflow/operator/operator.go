package operator

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidStatusToLaunch = errors.New("Cannot launch operator. The operator is in an invalid status.")

// Operator is an interface for managing and inspecting the lifecycle of an operator
// used by a workflow run.
type Operator interface {
	// Property getters. Retrieve property of the operator without making any changes.
	Type() operator.Type
	Name() string
	ID() uuid.UUID
	JobSpec() job.Spec
	// ExecState returns the operators ExecState since the last `Poll()` or `Launch()`
	ExecState() *shared.ExecutionState

	// Execution methods. These methods trigger or interact with jobs
	// and update operator's state.
	// However, they do not persist anything to DB.

	// Launch kicks off the execution of this operator, using operator's job spec.
	// It sets the operator's execState to 'Running' without updating DB.
	// Use `Poll()` afterwards to determine when this operator has completed.
	Launch(ctx context.Context) error

	// Poll performs a non-blocking fetch and update for the execution state of this operator.
	// Returns the execState updated. This does not persist the exec state to DB.
	Poll(ctx context.Context) (*shared.ExecutionState, error)

	// Cancel updates the status of this operator execution if the result of the
	// execution will not be generated. This does not persist the exec state to DB.
	Cancel()

	// Finish is an end-of-lifecycle hook meant to do any final cleanup work.
	// Also calls Finish() on all the operator's output artifacts.
	Finish(ctx context.Context)

	// DB methods, these methods update DB based on current operator's state.

	// InitializeResult initializes the operator in the database.
	// TODO: document.
	InitializeResult(ctx context.Context, dagResultID uuid.UUID) error

	// PersistResult writes the results of this operator execution to the database.
	// The result persisted is based on the last `Poll()`.
	//
	// Errors if the artifact hasn ot yet been computed, or InitializeResult() hasn't been called yet.
	// *This method also persists any artifact results produced by this operator.*
	PersistResult(ctx context.Context) error
}

// This should only be used within the boundaries of the execution engine.
// Specifies what we will do with the operator's results.
// Preview: *does not* persist workflow results or write to third-party integrations.
// Publish *does* both.
type ExecutionMode int

const (
	Preview ExecutionMode = iota
	Publish
)

func NewOperator(
	ctx context.Context,
	dbOperator operator.DBOperator,
	inputs []artifact.Artifact,
	outputs []artifact.Artifact,
	inputExecPaths []*utils.ExecPaths,
	outputExecPaths []*utils.ExecPaths,
	opResultWriter operator_result.Writer, // A nil value means the operator is run in preview mode.
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	previewCacheManager preview_cache.CacheManager,
	execMode ExecutionMode,
	db database.Database,
) (Operator, error) {
	resources := dbOperator.Spec.Resources()
	if resources != nil {
		log.Errorf("HELLO: Operator Spec Resources: %v", *resources)
		if resources.MemoryMB != nil {
			log.Errorf("HELLO: memory ", *resources.MemoryMB)
		}
		if resources.NumCpus != nil {
			log.Errorf("HELLO: num_cpus ", *resources.NumCpus)
		}
	}

	if len(inputs) != len(inputExecPaths) {
		return nil, errors.New("Internal error: mismatched number of input arguments.")
	}

	if len(outputs) != len(outputExecPaths) {
		return nil, errors.New("Internal error: mismatched number of input arguments.")
	}

	// If this operator has no outputs, we will need to allocate a new metadata path.
	// This is because the operator's metadata path is defined on the operator's outputs.
	metadataPath := uuid.New().String()
	if len(outputExecPaths) > 0 {
		metadataPath = outputExecPaths[0].OpMetadataPath
	}

	now := time.Now()

	baseOp := baseOperator{
		dbOperator:   &dbOperator,
		resultWriter: opResultWriter,
		resultID:     uuid.Nil,

		metadataPath: metadataPath,
		jobName:      "", /* Must be set by the specific type constructors below. */

		inputs:          inputs,
		outputs:         outputs,
		inputExecPaths:  inputExecPaths,
		outputExecPaths: outputExecPaths,

		previewCacheManager: previewCacheManager,
		jobManager:          jobManager,
		vaultObject:         vaultObject,
		storageConfig:       storageConfig,
		db:                  db,

		execMode: execMode,
		execState: shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
			Timestamps: &shared.ExecutionTimestamps{
				PendingAt: &now,
			},
		},

		// These fields may be set dynamically during orchestration.
		resultsPersisted: false,
	}

	if dbOperator.Spec.IsFunction() {
		return newFunctionOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsMetric() {
		return newMetricOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsCheck() {
		return newCheckOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsExtract() {
		return newExtractOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsLoad() {
		return newLoadOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsParam() {
		return newParamOperator(baseOp)
	} else if dbOperator.Spec.IsSystemMetric() {
		return newSystemMetricOperator(baseOp)
	}

	return nil, errors.Newf("Unsupported operator type %s", dbOperator.Spec.Type())
}
