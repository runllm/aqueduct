package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type artifactResultRepo struct {
	artifactResultReader
	artifactResultWriter
}

type artifactResultReader struct{}

type artifactResultWriter struct{}

func NewArtifactResultRepo() repos.ArtifactResult {
	return &artifactResultRepo{
		artifactResultReader: artifactResultReader{},
		artifactResultWriter: artifactResultWriter{},
	}
}

func (*artifactResultReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE id = $1`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{ID}

	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE id IN (%s);`,
		models.ArtifactResultCols(),		
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifact(ctx context.Context, artifactID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE artifact_id = $1;`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{artifactID}
	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifactAndWorkflow(ctx context.Context, workflowID uuid.UUID, artifactName string, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT %s FROM artifact_result, artifact, workflow_dag, workflow_dag_edge
		WHERE workflow_dag.workflow_id = $1
		AND artifact.name = $2
		AND (
			workflow_dag_edge.from_id = artifact.id
			OR
			workflow_dag_edge.to_id = artifact.id
		)
		AND workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND artifact_result.artifact_id = artifact.id;`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{workflowID, artifactName}
	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifactAndDAGResult(ctx context.Context, dagResultID uuid.UUID, artifactID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE workflow_dag_result_id = $1 AND artifact_id = $2;`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{dagResultID, artifactID}
	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByDAGResults(ctx context.Context, dagResultIDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) { 
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE workflow_dag_result_id IN (%s);`,
		models.ArtifactResultCols(),
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)

	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultWriter) Create(
	ctx context.Context,
	dagResultID uuid.UUID,
	artifactID uuid.UUID,
	contentPath string,
	DB database.Database,
) (*models.ArtifactResult, error) {
	cols := []string{
		models.ArtifactResultID,
		models.ArtifactResultDAGResultID,
		models.ArtifactResultArtifactID,
		models.ArtifactResultContentPath,
		models.ArtifactResultStatus,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ArtifactResultTable, cols, models.ArtifactResultCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.ArtifactResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagResultID,
		artifactID,
		contentPath,
		shared.PendingExecutionStatus,
	}
	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultWriter) CreateWithExecStateAndMetadata(
	ctx context.Context,
	dagResultID uuid.UUID,
	artifactID uuid.UUID,
	contentPath string,
	execState *shared.ExecutionState,
	metadata *shared.ArtifactResultMetadata,
	DB database.Database,
) (*models.ArtifactResult, error) {
	cols := []string{
		models.ArtifactResultID,
		models.ArtifactResultDAGResultID,
		models.ArtifactResultArtifactID,
		models.ArtifactResultContentPath,
		models.ArtifactResultStatus,
		models.ArtifactResultStatus,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ArtifactResultTable, cols, models.ArtifactResultCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.ArtifactResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagResultID,
		artifactID,
		contentPath,
		shared.PendingExecutionStatus,
	}
	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteArtifactResults(ctx, DB, []uuid.UUID{ID})
}

func (*artifactResultWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error { 
	return deleteArtifactResults(ctx, DB, IDs)
}

func (*artifactResultWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ArtifactResult, error) {
	var artifact_result models.ArtifactResult
	err := utils.UpdateRecordToDest(
		ctx,
		&artifact_result,
		changes,
		models.ArtifactResultTable,
		models.ArtifactResultID,
		ID,
		models.ArtifactCols(),
		DB,
	)
	return &artifact_result, err
}

func deleteArtifactResults(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM artifact_result WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}

func getArtifactResults(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.ArtifactResult, error) {
	var artifact_results []models.ArtifactResult
	err := DB.Query(ctx, &artifact_results, query, args...)
	return artifact_results, err
}

func getArtifactResult(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.ArtifactResult, error) {
	artifact_results, err := getArtifactResults(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(artifact_results) == 0 {
		return nil, database.ErrNoRows
	}

	if len(artifact_results) != 1 {
		return nil, errors.Newf("Expected 1 artifact_result but got %v", len(artifact_results))
	}

	return &artifact_results[0], nil
}
