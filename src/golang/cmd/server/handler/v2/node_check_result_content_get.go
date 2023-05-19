package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodeCheckResultContentGet.tsx
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/check/{nodeID}/result/{nodeResultID}/content
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
//	`nodeID`: ID for node object
//	`nodeResultID`: ID for node result object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		`response.NodeContent`

type NodeCheckResultContentGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo       repos.Workflow
	DAGRepo            repos.DAG
	OperatorRepo       repos.Operator
	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
}

func (*NodeCheckResultContentGetHandler) Name() string {
	return "NodeCheckResultContentGet"
}

func (h *NodeCheckResultContentGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dagID, err := (parser.DagIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	nodeID, err := (parser.NodeIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	nodeResultID, err := (parser.NodeResultIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &nodeResultGetArgs{
		AqContext:    aqContext,
		workflowID:   workflowID,
		dagID:        dagID,
		nodeID:       nodeID,
		nodeResultID: nodeResultID,
	}, http.StatusOK, nil
}

func (h *NodeCheckResultContentGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodeResultGetArgs)
	emptyResp := &nodeResultGetResponse{}

<<<<<<< HEAD
	dag, err := h.DAGRepo.GetByDAGResult(
		ctx,
		args.nodeResultID,
=======
	dag, err := h.DAGRepo.Get(
		ctx,
		args.dagID,
>>>>>>> 52ad1258eb8084018de394390c4cc2250e14ed57
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow dag.")
	}

	dbOperatorWithArtifactNode, err := h.OperatorRepo.GetOperatorWithArtifactNode(ctx, args.nodeID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading check node.")
	}

	dbArtifact, err := h.ArtifactRepo.Get(ctx, dbOperatorWithArtifactNode.ArtifactID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact result.")
	}

	execState := shared.ExecutionState{}
	dbArtifactResult, err := h.ArtifactResultRepo.Get(
		ctx,
		args.nodeResultID,
		h.Database,
	)
	if err != nil {
		if !errors.Is(err, database.ErrNoRows()) {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact result.")
		}
		// ArtifactResult was never created, so we mark the artifact as cancelled.
		execState.Status = shared.CanceledExecutionStatus
	} else {
		execState.Status = dbArtifactResult.Status
	}

	// `dbArtifactResult` is not guaranteed to be non-nil here.
	if dbArtifactResult != nil && !dbArtifactResult.ExecState.IsNull {
		execState.FailureType = dbArtifactResult.ExecState.FailureType
		execState.Error = dbArtifactResult.ExecState.Error
		execState.UserLogs = dbArtifactResult.ExecState.UserLogs
	}

	artifactObject := artifact.NewArtifactFromDBObjects(
		uuid.UUID{}, /* signature */
		dbArtifact,
		dbArtifactResult,
		h.ArtifactRepo,
		h.ArtifactResultRepo,
		&dag.StorageConfig,
		nil, /* previewCacheManager */
		h.Database,
	)

	data, isDownsampled, err := artifactObject.SampleContent(ctx)
	// Should not be downsampled.
	if err != nil || isDownsampled {
		if errors.Is(err, storage.ErrObjectDoesNotExist()) {
			return emptyResp, http.StatusOK, nil
		}

		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve data for the artifact result.")
<<<<<<< HEAD
	} else if !errors.Is(err, storage.ErrObjectDoesNotExist()) {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve data for the artifact result.")
	}

	return &nodeResultGetResponse{IsDownsampled: isDownsampled, Content: data}, http.StatusOK, nil
}
=======
	}

	return &nodeResultGetResponse{IsDownsampled: isDownsampled, Content: data}, http.StatusOK, nil
}
>>>>>>> 52ad1258eb8084018de394390c4cc2250e14ed57
