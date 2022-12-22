package databricks

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib"
	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/clusters"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/databricks/databricks-sdk-go/service/libraries"
	"github.com/dropbox/godropbox/errors"
)

func NewWorkspaceClient(
	workspaceUrl string,
	accessToken string,
) (*databricks.WorkspaceClient, error) {
	dConfig := &databricks.Config{
		Host:  workspaceUrl,
		Token: accessToken,
	}
	datatbricksClient, err := databricks.NewWorkspaceClient(dConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create Databricks client.")
	}
	return datatbricksClient, nil
}

func ListJobs(
	ctx context.Context,
	databricksClient *databricks.WorkspaceClient,
) ([]jobs.Job, error) {
	jobs, err := databricksClient.Jobs.ListAll(
		ctx,
		jobs.List{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error launching job in Databricks.")
	}
	return jobs, nil
}

func CreateJob(
	ctx context.Context,
	databricksClient *databricks.WorkspaceClient,
	name string,
	s3InstanceProfileArn string,
	pythonFilePath string,
) (int64, error) {
	createRequest := &jobs.CreateJob{
		Name: name,
		Tasks: []jobs.JobTaskSettings{
			{
				TaskKey: name,
				NewCluster: &clusters.CreateCluster{
					SparkVersion: SparkVersion,
					NumWorkers:   NumWorkers,
					NodeTypeId:   NodeTypeId,
					AwsAttributes: &clusters.AwsAttributes{
						InstanceProfileArn: s3InstanceProfileArn,
					},
				},
				SparkPythonTask: &jobs.SparkPythonTask{
					PythonFile: pythonFilePath,
				},
				Libraries: []libraries.Library{
					{
						Pypi: &libraries.PythonPyPiLibrary{
							Package: fmt.Sprintf("aqueduct-ml==%s", lib.ServerVersionNumber),
						},
					},
				},
			},
		},
	}
	createResp, err := databricksClient.Jobs.Create(ctx, *createRequest)
	if err != nil {
		return -1, errors.Wrap(err, "Error creating a job in Databricks.")
	}
	return createResp.JobId, nil
}

func RunNow(
	ctx context.Context,
	databricksClient *databricks.WorkspaceClient,
	jobId int64,
	specStr string,
) (int64, error) {
	runResp, err := databricksClient.Jobs.RunNow(
		ctx,
		jobs.RunNow{
			JobId:        jobId,
			PythonParams: []string{"--spec", specStr},
		},
	)
	if err != nil {
		return -1, errors.Wrap(err, "Error launching job in Databricks.")
	}
	return runResp.RunId, nil
}
