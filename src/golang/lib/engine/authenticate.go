package engine

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/k8s"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const MaxConcurrentDownload = 3

// Authenticates kubernetes configuration by trying to connect a client.
func AuthenticateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := lib_utils.ParseK8sConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}
	_, err = k8s.CreateK8sClient(conf.KubeconfigPath, bool(conf.UseSameCluster))
	if err != nil {
		return errors.Wrap(err, "Unable to create kubernetes client.")
	}
	return nil
}

func AuthenticateLambdaConfig(ctx context.Context, authConf auth.Config) error {
	lambdaConf, err := lib_utils.ParseLambdaConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	functionsToShip := [10]lambda_utils.LambdaFunctionType{
		lambda_utils.FunctionExecutor37Type,
		lambda_utils.FunctionExecutor38Type,
		lambda_utils.FunctionExecutor39Type,
		lambda_utils.ParamExecutorType,
		lambda_utils.SystemMetricType,
		lambda_utils.AthenaConnectorType,
		lambda_utils.BigQueryConnectorType,
		lambda_utils.PostgresConnectorType,
		lambda_utils.S3ConnectorType,
		lambda_utils.SnowflakeConnectorType,
	}

	errGroup, _ := errgroup.WithContext(ctx)
	lambda_utils.AuthenticateDockerToECR()

	for i := 0; i < len(functionsToShip); i += MaxConcurrentDownload {
		for j := 0; j < MaxConcurrentDownload; j++ {
			if j+i < len(functionsToShip) {
				log.Info(j)
				lambdaFunctionType := functionsToShip[j+i]
				errGroup.Go(func() error {
					return lambda_utils.PullImageFromECR(lambdaFunctionType)
				})
			}
		}
		if err := errGroup.Wait(); err != nil {
			return errors.Wrap(err, "Unable to Create Lambda Function.")
		}
	}

	for _, functionType := range functionsToShip {
		lambdaFunctionType := functionType
		errGroup.Go(func() error {
			return lambda_utils.CreateLambdaFunction(lambdaFunctionType, lambdaConf.RoleArn)
		})
	}

	if err := errGroup.Wait(); err != nil {
		return errors.Wrap(err, "Unable to Create Lambda Function.")
	}

	return nil
}
