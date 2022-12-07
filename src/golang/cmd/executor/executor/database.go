package executor

import (
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Readers struct {
	OperatorReader             operator.Reader
	IntegrationReader          integration.Reader
	OperatorResultReader       operator_result.Reader
	ExecutionEnvironmentReader exec_env.Reader
}

type Writers struct {
	WorkflowWatcherWriter workflow_watcher.Writer
	OperatorWriter        operator.Writer
	OperatorResultWriter  operator_result.Writer
}

type Repos struct {
	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
	DAGEdgeRepo        repos.DAGEdge
	DAGResultRepo      repos.DAGResult
	NotificationRepo   repos.Notification
	WatcherRepo        repos.Watcher
	WorkflowRepo       repos.Workflow
}

func createReaders(dbConf *database.DatabaseConfig) (*Readers, error) {
	operatorReader, err := operator.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	integrationReader, err := integration.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	operatorResultReader, err := operator_result.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	execEnvReader, err := exec_env.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	return &Readers{
		OperatorReader:             operatorReader,
		IntegrationReader:          integrationReader,
		OperatorResultReader:       operatorResultReader,
		ExecutionEnvironmentReader: execEnvReader,
	}, nil
}

func createWriters(dbConf *database.DatabaseConfig) (*Writers, error) {
	workflowWatcherWriter, err := workflow_watcher.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	operatorWriter, err := operator.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	operatorResultWriter, err := operator_result.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	return &Writers{
		WorkflowWatcherWriter: workflowWatcherWriter,
		OperatorWriter:        operatorWriter,
		OperatorResultWriter:  operatorResultWriter,
	}, nil
}

func createRepos() *Repos {
	return &Repos{
		ArtifactRepo:       sqlite.NewArtifactRepo(),
		ArtifactResultRepo: sqlite.NewArtifactResultRepo(),
		DAGRepo:            sqlite.NewDAGRepo(),
		DAGEdgeRepo:        sqlite.NewDAGEdgeRepo(),
		DAGResultRepo:      sqlite.NewDAGResultRepo(),
		NotificationRepo:   sqlite.NewNotificationRepo(),
		WatcherRepo:        sqlite.NewWatcherRepo(),
		WorkflowRepo:       sqlite.NewWorklowRepo(),
	}
}

func getEngineReaders(readers *Readers) *engine.EngineReaders {
	return &engine.EngineReaders{
		OperatorReader:             readers.OperatorReader,
		OperatorResultReader:       readers.OperatorResultReader,
		IntegrationReader:          readers.IntegrationReader,
		ExecutionEnvironmentReader: readers.ExecutionEnvironmentReader,
	}
}

func getEngineWriters(writers *Writers) *engine.EngineWriters {
	return &engine.EngineWriters{
		OperatorWriter:       writers.OperatorWriter,
		OperatorResultWriter: writers.OperatorResultWriter,
	}
}

func getEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		ArtifactRepo:       repos.ArtifactRepo,
		ArtifactResultRepo: repos.ArtifactResultRepo,
		DAGRepo:            repos.DAGRepo,
		DAGEdgeRepo:        repos.DAGEdgeRepo,
		DAGResultRepo:      repos.DAGResultRepo,
		NotificationRepo:   repos.NotificationRepo,
		WatcherRepo:        repos.WatcherRepo,
		WorkflowRepo:       repos.WorkflowRepo,
	}
}
