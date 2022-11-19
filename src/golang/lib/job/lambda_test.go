package job

import (
	"context"
	"fmt"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLambdaAPI(t *testing.T) {
	//t.Skip( " ERROR MESSAGE" ) // TODO:

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	lambdaSvc := lambda.New(sess)

	functionName := lambda_utils.FunctionLambdaFunction38

	jobManager := &lambdaJobManager{
		lambdaService: lambdaSvc,
	}

	newMemory := int64(300)
	oldMemory, err := jobManager.updateFunctionMemory(context.Background(), functionName, &newMemory)
	require.Nil(t, err)
	fmt.Println("OLD MEMORY: ", *oldMemory)
}
