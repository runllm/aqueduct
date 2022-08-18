package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	metadataFormFieldName = "metadata"
	dataFormFieldName     = "data"
)

// Route: /artifact_result/{workflowDagResultId}/{artifactId}
// Method: GET
// Params:
//
//	`workflowDagResultId`: ID for `workflow_dag_result` object
//	`artifactId`: ID for `artifact` object
//
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response:
//
//	Body:
//		serialized `getArtifactResultResponse`,
//		metadata and content of the result of `artifactId` on the given workflow_dag_result object.
type getArtifactResultArgs struct {
	*aq_context.AqContext
	workflowDagResultId uuid.UUID
	artifactId          uuid.UUID
}

type artifactResultMetadata struct {
	Name string `json:"name"`

	// `Status` is redundant due to `ExecState`. Avoid consuming `Status` in new code.
	// We are incurring this tech debt right now since there are quite a few usages of
	// `status` in the UI.
	Status            shared.ExecutionStatus            `json:"status"`
	ExecState         shared.ExecutionState             `json:"exec_state"`
	Schema            []map[string]string               `json:"schema"`
	SerializationType artifact_result.SerializationType `json:"serialization_type"`
	ArtifactType      artifact.Type                     `json:"artifact_type"`
}

type getArtifactResultResponse struct {
	Metadata *artifactResultMetadata `json:"metadata"`
	Data     []byte                  `json:"data"`
}

type GetArtifactResultHandler struct {
	GetHandler

	Database             database.Database
	ArtifactReader       artifact.Reader
	ArtifactResultReader artifact_result.Reader
	WorkflowDagReader    workflow_dag.Reader
}

func (*GetArtifactResultHandler) Name() string {
	return "GetArtifactResult"
}

func (*GetArtifactResultHandler) SendResponse(w http.ResponseWriter, response interface{}) {
	resp := response.(*getArtifactResultResponse)
	multipartWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", multipartWriter.FormDataContentType())

	metadataJsonBlob, err := json.Marshal(resp.Metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formFieldWriter, err := multipartWriter.CreateFormField(metadataFormFieldName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = formFieldWriter.Write(metadataJsonBlob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Data) > 0 {
		formFieldWriter, err = multipartWriter.CreateFormField(dataFormFieldName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = formFieldWriter.Write(resp.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = multipartWriter.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *GetArtifactResultHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowDagResultIdStr := chi.URLParam(r, routes.WorkflowDagResultIdUrlParam)
	workflowDagResultId, err := uuid.Parse(workflowDagResultIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow dag result ID.")
	}

	artifactIdStr := chi.URLParam(r, routes.ArtifactIdUrlParam)
	artifactId, err := uuid.Parse(artifactIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed artifact ID.")
	}

	ok, err := h.ArtifactReader.ValidateArtifactOwnership(
		r.Context(),
		aqContext.OrganizationId,
		artifactId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during artifact ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this artifact.")
	}

	return &getArtifactResultArgs{
		AqContext:           aqContext,
		workflowDagResultId: workflowDagResultId,
		artifactId:          artifactId,
	}, http.StatusOK, nil
}

func (h *GetArtifactResultHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getArtifactResultArgs)

	emptyResp := getArtifactResultResponse{}

	workflowDag, err := h.WorkflowDagReader.GetWorkflowDagByWorkflowDagResultId(
		ctx,
		args.workflowDagResultId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow dag.")
	}

	dbArtifact, err := h.ArtifactReader.GetArtifact(ctx, args.artifactId, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact result.")
	}

	dbArtifactResult, err := h.ArtifactResultReader.GetArtifactResultByWorkflowDagResultIdAndArtifactId(
		ctx,
		args.workflowDagResultId,
		args.artifactId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact result.")
	}

	execState := shared.ExecutionState{
		Status: dbArtifactResult.Status,
	}
	if !dbArtifactResult.ExecState.IsNull {
		execState.FailureType = dbArtifactResult.ExecState.FailureType
		execState.Error = dbArtifactResult.ExecState.Error
		execState.UserLogs = dbArtifactResult.ExecState.UserLogs
	}

	metadata := artifactResultMetadata{
		Status:            execState.Status,
		ExecState:         execState,
		Name:              dbArtifact.Name,
		ArtifactType:      dbArtifactResult.Metadata.ArtifactType,
		SerializationType: dbArtifactResult.Metadata.SerializationType,
	}

	if !dbArtifactResult.Metadata.IsNull {
		metadata.Schema = dbArtifactResult.Metadata.Schema
	}

	response := &getArtifactResultResponse{
		Metadata: &metadata,
	}

	if dbArtifactResult.Status == shared.SucceededExecutionStatus {
		// We retrieve the data only when the artifact result status is `succeeded`.
		data, err := storage.NewStorage(&workflowDag.StorageConfig).Get(
			ctx,
			dbArtifactResult.ContentPath,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve data for the artifact result.")
		}

		response.Data = data
	}

	return response, http.StatusOK, nil
}
