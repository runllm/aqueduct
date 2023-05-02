package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/IntegrationsWorkflowsGet.ts
//
// Route: /v2/integrations/workflows
// Method: GET
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		Map of integration ID to list of workflow IDs that use that integration.

type integrationsWorkflowsGetArgs struct {
	*aq_context.AqContext
}

type IntegrationsWorkflowsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	OperatorRepo    repos.Operator
}

func (*IntegrationsWorkflowsGetHandler) Name() string {
	return "IntegrationsWorkflowsGet"
}

func (h *IntegrationsWorkflowsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &integrationsWorkflowsGetArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *IntegrationsWorkflowsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*integrationsWorkflowsGetArgs)

	integrations, err := h.IntegrationRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list integrations.")
	}

	response := make(map[uuid.UUID][]uuid.UUID, len(integrations))
	for _, integration := range integrations {
		workflowIDs, err := fetchWorkflowIDsForIntegration(ctx, args.OrgID, &integration, h.IntegrationRepo, h.OperatorRepo, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find workflows for integration %s", integration.ID)
		}
		response[integration.ID] = workflowIDs
	}
	return response, http.StatusOK, nil
}

// fetchWorkflowIDsForIntegration returns a list of workflow IDs that use the given integration.
func fetchWorkflowIDsForIntegration(
	ctx context.Context,
	orgID string,
	integration *models.Integration,
	integrationRepo repos.Integration,
	operatorRepo repos.Operator,
	db database.Database,
) ([]uuid.UUID, error) {
	operators, err := operator.GetOperatorsOnIntegration(
		ctx,
		orgID,
		integration,
		integrationRepo,
		operatorRepo,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operators.")
	}

	// Now, using the operators using this integration, we can infer all the workflows
	// that also use this integration.
	operatorIDs := slices.Map(operators, func(op models.Operator) uuid.UUID {
		return op.ID
	})

	operatorRelations, err := operatorRepo.GetRelationBatch(ctx, operatorIDs, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}

	workflowIDSet := make(map[uuid.UUID]bool, len(operatorRelations))
	workflowIDs := make([]uuid.UUID, 0, len(operatorRelations))
	for _, operatorRelation := range operatorRelations {
		if _, ok := workflowIDSet[operatorRelation.WorkflowID]; ok {
			continue
		}
		workflowIDSet[operatorRelation.WorkflowID] = true
		workflowIDs = append(workflowIDs, operatorRelation.WorkflowID)
	}
	return workflowIDs, nil
}