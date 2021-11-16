package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go/aws"
)

const (
	AWS_REGION     = "ap-south-1"
	FORMAT_VERSION = "aws_v1"
)

var (
	client *pricing.Client
)

type AwsService struct {
	Context     context.Context
	ServiceCode string
}

func Configure(ctx context.Context) error {
	// Configuration for AWS
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(AWS_REGION))
	if err != nil {
		return err
	}
	// Create client
	client = pricing.NewFromConfig(cfg)
	return nil
}

func NewService(ctx context.Context, serviceCode string) *AwsService {
	return &AwsService{Context: ctx, ServiceCode: serviceCode}
}

func (as AwsService) GetAttributes() ([]string, error) {
	// Set input parameter
	input := &pricing.DescribeServicesInput{
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(1),
		ServiceCode:   aws.String(as.ServiceCode),
	}
	// Execute command
	output, err := client.DescribeServices(as.Context, input)
	if err != nil {
		return nil, err
	} else if len(output.Services) == 0 {
		return nil, errors.New("Not found service for service code")
	} else {
		return output.Services[0].AttributeNames, nil
	}
}

func (as AwsService) GetAttributeValues(attribute string) ([]string, error) {
	// Set input parameter
	input := &pricing.GetAttributeValuesInput{
		AttributeName: aws.String(attribute),
		MaxResults:    int32(100),
		ServiceCode:   aws.String(as.ServiceCode),
	}
	// Create paginator
	paginator := pricing.NewGetAttributeValuesPaginator(client, input)
	// Extract result
	result := make([]string, 0)
	for {
		// Execute query
		output, err := paginator.NextPage(as.Context)
		if err != nil {
			return nil, err
		}
		// Extract
		for _, value := range output.AttributeValues {
			result = append(result, *value.Value)
		}
		// Escape
		if !paginator.HasMorePages() {
			return result, nil
		}
	}
}