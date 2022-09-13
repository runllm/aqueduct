package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type postgresReaderImpl struct {
	standardReaderImpl
}

type postgresWriterImpl struct {
	standardWriterImpl
}

func newPostgresReader() Reader {
	return &postgresReaderImpl{standardReaderImpl{}}
}

func newPostgresWriter() Writer {
	return &postgresWriterImpl{standardWriterImpl{}}
}

func (r *postgresReaderImpl) GetLoadOperatorsForWorkflowAndIntegration(
	ctx context.Context,
	workflowId uuid.UUID,
	integrationId uuid.UUID,
	objectName string,
	db database.Database,
) ([]DBOperator, error) {
	// Get all load operators where table=objectName & integration_id=integrationId
	// and has an edge (in `from_id` or `to_id`) in a DAG belonging to the specified
	// workflow.
	query := fmt.Sprintf(`
	SELECT %s
	FROM operator
	WHERE
		json_extract_path_text(spec, 'type') = '%s' AND 
		json_extract_path_text(spec, 'load', 'parameters', 'table')=$1 AND
		json_extract_path_text(spec, 'load', 'integration_id')=$2 AND
		EXISTS (
			SELECT 1 
			FROM 
				workflow_dag_edge, workflow_dag 
			WHERE 
				( 
					workflow_dag_edge.from_id = operator.id OR 
					workflow_dag_edge.to_id = operator.id 
				) AND 
				workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
				workflow_dag.workflow_id = $4
		);`, allColumns(), LoadType)

	var operators []DBOperator
	err := db.Query(ctx, &operators, query, objectName, integrationId, workflowId)

	return operators, err
}

func (r *postgresReaderImpl) GetOperatorsByIntegrationId(
	ctx context.Context,
	integrationId uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	getOperatorsByIntegrationIdQuery := fmt.Sprintf(
		`SELECT %s FROM %s
		WHERE json_extract_text(spec, 'load', 'integration_id') = $1
		OR json_extract_text(spec, 'extract', 'integration_id') = $2`,
		allColumns(),
		tableName,
	)

	var operators []DBOperator
	err := db.Query(
		ctx,
		&operators,
		getOperatorsByIntegrationIdQuery,
		integrationId,
		integrationId,
	)
	return operators, err
}

func (r *postgresReaderImpl) GetDistinctLoadOperatorsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]GetDistinctLoadOperatorsByWorkflowIdResponse, error) {
	// Get all unique load operator (defined as a unique combination of integration,
	// table, and update mode) that has an edge (in `from_id` or `to_id`) in a DAG
	// belonging to the specified workflow in order of when the operator was last modified.
	query := `
	SELECT DISTINCT ON
		(
			integration.name AS integration_name, 
			json_extract_path_text(operator.spec, 'load', 'integration_id') AS integration_id,
			json_extract_path_text(operator.spec, 'load', 'service') AS service, 
			json_extract_path_text(operator.spec, 'load', 'parameters', 'table') AS table_name, 
			json_extract_path_text(operator.spec, 'load', 'parameters', 'update_mode')  AS update_mode
		)
		operator.name AS operator_name,
		workflow_dag.created_at AS modified_at,
	FROM 
		operator, integration, workflow_dag_edge, workflow_dag  
	WHERE (
		json_extract_path_text(spec, 'type') = 'load' AND 
		integration.id = json_extract_path_text(operator.spec, 'load', 'integration_id') AND
		( 
			workflow_dag_edge.from_id = operator.id OR 
			workflow_dag_edge.to_id = operator.id 
		) AND 
		workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = $1
	)
	ORDER BY modified_at DESC;`

	var workflowSpecs []GetDistinctLoadOperatorsByWorkflowIdResponse
	err := db.Query(ctx, &workflowSpecs, query, workflowId)
	return workflowSpecs, err
}
