package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/orbit-ops/launchpad-core/ent"
	"github.com/orbit-ops/launchpad-core/providers"
)

type AwsProvider struct {
	providers.BaseProvider

	lambdas *lambda.Client
	conf    *providers.ProviderConfig
}

func NewAwsProvider(c *providers.ProviderConfig) (*AwsProvider, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	return &AwsProvider{
		lambdas: lambda.NewFromConfig(cfg),
		conf:    c,
	}, nil
}

func (ap *AwsProvider) ScheduleDeletion() error {
	return nil
}

func (ap *AwsProvider) CreateAccess(ctx context.Context, token string, rocket *ent.Rocket, req *ent.Request) error {
	return ap.runLambda(ctx, providers.CreateAccess, token, rocket, req)
}

func (ap *AwsProvider) RemoveAccess(ctx context.Context, token string, rocket *ent.Rocket, req *ent.Request) error {
	return ap.runLambda(ctx, providers.RemoveAccess, token, rocket, req)
}

func (ap *AwsProvider) runLambda(ctx context.Context, command providers.ProviderCommand, token string, rocket *ent.Rocket, req *ent.Request) error {
	rc, err := ap.EncodeRocketConfig(req)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"command": command,
		"config":  rc,
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	// Prepare input for invoking Lambda function
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(fmt.Sprintf("launchpad-access-%s", req.ID)),
		Payload:        payloadBytes,
		InvocationType: "Event",
	}

	// Invoke Lambda function
	result, err := ap.lambdas.Invoke(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("invoking lambda: %w", err)
	}

	fmt.Println("Lambda function invoked successfully.")
	fmt.Println("Response:", string(result.Payload))

	return nil
}
