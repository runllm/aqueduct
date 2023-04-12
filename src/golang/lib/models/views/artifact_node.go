package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	ArtifactNodeView    = "artifact_node"
	ArtifactNodeDagID   = "dag_id"
	ArtifactNodeInputs  = "inputs"
	ArtifactNodeOutputs = "outputs"

	ArtifactNodeViewSubQuery = `
	WITH artf_with_outputs AS ( -- Aggregate outputs
		SELECT
			artifact.id AS id,
			workflow_dag.id AS dag_id,
			artifact.name AS name,
			artifact.description AS description,
			artifact.type as type,
			CAST( json_group_array(
				json_object(
					'value', workflow_dag_edge.to_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS outputs
		FROM
			artifact, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND artifact.id = workflow_dag_edge.from_id
		GROUP BY
			workflow_dag.id, artifact.id
	),
	artf_with_inputs AS ( -- Aggregate inputs
		SELECT
			artifact.id AS id,
			workflow_dag.id AS dag_id,
			artifact.name AS name,
			artifact.description AS description,
			artifact.type as type,
			CAST( json_group_array(
				json_object(
					'value', workflow_dag_edge.from_id,
					'idx', workflow_dag_edge.idx
				)
			) AS BLOB) AS inputs
		FROM
			artifact, workflow_dag, workflow_dag_edge
		WHERE
			workflow_dag.id = workflow_dag_edge.workflow_dag_id
			AND artifact.id = workflow_dag_edge.to_id
		GROUP BY
			workflow_dag.id, artifact.id
	)
	SELECT -- just do input LEFT JOIN outputs as all artifacts have inputs
		artf_with_inputs.id AS id,
		artf_with_inputs.dag_id AS dag_id,
		artf_with_inputs.name AS name,
		artf_with_inputs.description AS description,
		artf_with_inputs.type AS type,
		artf_with_outputs.outputs AS outputs,
		artf_with_inputs.inputs AS inputs
	FROM
		artf_with_inputs LEFT JOIN artf_with_outputs
	ON
		artf_with_outputs.id = artf_with_inputs.id
		AND artf_with_outputs.dag_id = artf_with_inputs.dag_id
	`
)

type ArtifactNode struct {
	ID          uuid.UUID           `db:"id" json:"id"`
	DagID       uuid.UUID           `db:"dag_id" json:"dag_id"`
	Name        string              `db:"name" json:"name"`
	Description string              `db:"description" json:"description"`
	Type        shared.ArtifactType `db:"type" json:"type"`

	Inputs  shared.NullableIndexedList[uuid.UUID] `db:"inputs" json:"inputs"`
	Outputs shared.NullableIndexedList[uuid.UUID] `db:"outputs" json:"outputs"`
}

// ArtifactNodeCols returns a comma-separated string of all artifact columns.
func ArtifactNodeCols() string {
	return strings.Join(allArtifactNodeCols(), ",")
}

// ArtifactNodeColsWithPrefix returns a comma-separated string of all
// artifact columns prefixed by the view name.
func ArtifactNodeColsWithPrefix() string {
	cols := allArtifactNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", ArtifactNodeView, col)
	}

	return strings.Join(cols, ",")
}

func allArtifactNodeCols() []string {
	artfNodeCols := models.AllArtifactCols()
	artfNodeCols = append(
		artfNodeCols,
		ArtifactNodeDagID,
		ArtifactNodeInputs,
		ArtifactNodeOutputs,
	)

	return artfNodeCols
}