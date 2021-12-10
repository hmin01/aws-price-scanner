package pricing

import (
	"context"
	"errors"

	// AWS
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsPricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"

	// Model
	"aws-price-scanner/model"
	// Process
	"aws-price-scanner/process"
)

const (
	AWS_REGION     = "ap-south-1"
	FORMAT_VERSION = "aws_v1"
)

var svc *awsPricing.Client

type AwsService struct {
	Context     context.Context
	ServiceCode string
}

/*
 * AWS pricing configuration
 * @param 		ctx {context.Context} context
 * @response	{error} error object (contain nil)
 */
func Configure(ctx context.Context) error {
	// Configuration for AWS
	if cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(AWS_REGION)); err != nil {
		return err
	} else {
		// Create service client for aws pricing
		svc = awsPricing.NewFromConfig(cfg)
		return nil
	}
}

/*
 * Get a list of service code to use AWS pricing SDK
 * @param 		ctx {context.Context} context
 * @response	{[]string} a list of service code
 * @response	{error} error object (contain nil)
 */
func GetServiceCodeList(ctx context.Context) ([]string, error) {
	// Set input parameter
	input := &awsPricing.DescribeServicesInput{
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(100),
	}
	// Create paginator
	paginator := awsPricing.NewDescribeServicesPaginator(svc, input)
	// Extract service code from result
	result := make([]string, 0)
	for {
		// Execute command and extract service code
		if output, err := paginator.NextPage(ctx); err != nil {
			return nil, err
		} else {
			for _, value := range output.Services {
				result = append(result, *value.ServiceCode)
			}
		}
		// Escape logic
		if !paginator.HasMorePages() {
			return result, nil
		}
	}
}

/*
 * New service object to use AWS pricing SDK
 * @param 		ctx {context.Context} context
 * @param			serviceCode {string} service code
 * @response	{*AwsService} service object
 */
func NewService(ctx context.Context, serviceCode string) *AwsService {
	return &AwsService{Context: ctx, ServiceCode: serviceCode}
}

/*
 * [Method] Get a list of service attribute
 * @response	{[]string} a list of service attribute
 * @response 	{error} error object (contain nil)
 */
func (as AwsService) GetAttributes() ([]string, error) {
	// Set input parameter
	input := &awsPricing.DescribeServicesInput{
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(1),
		ServiceCode:   aws.String(as.ServiceCode),
	}
	// Execute command
	if output, err := svc.DescribeServices(as.Context, input); err != nil {
		return nil, err
	} else if len(output.Services) == 0 {
		return nil, errors.New("Not found service for service code")
	} else {
		return output.Services[0].AttributeNames, nil
	}
}

/*
 * [Method] Get a list of attribute value for service attribute
 * @param			attribute {string} attribute
 * @response	{[]string} a list of attribute value
 * @response 	{error} error object (contain nil)
 */
func (as AwsService) GetAttributeValues(attribute string) ([]string, error) {
	// Set input parameter
	input := &awsPricing.GetAttributeValuesInput{
		AttributeName: aws.String(attribute),
		MaxResults:    int32(100),
		ServiceCode:   aws.String(as.ServiceCode),
	}
	// Create paginator
	paginator := awsPricing.NewGetAttributeValuesPaginator(svc, input)
	// Extract service code from result
	result := make([]string, 0)
	for {
		// Execute command and extract attribute values
		if output, err := paginator.NextPage(as.Context); err != nil {
			return nil, err
		} else {
			for _, value := range output.AttributeValues {
				result = append(result, *value.Value)
			}
		}
		// Escape logic
		if !paginator.HasMorePages() {
			return result, nil
		}
	}
}

/*
 * [Method] Get a list of price information for service (store output in AWS S3)
 */
func (as AwsService) GetPriceList() {
	// Set filters
	var filters []types.Filter
	switch as.ServiceCode {
	case model.AWS_SERVICE_CODE_DYNAMODB:
		filters = []types.Filter{{
			Field: aws.String("termType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("OnDemand"),
		}}
	case model.AWS_SERVICE_CODE_EBS:
		filters = []types.Filter{{
			Field: aws.String("productFamily"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Storage"),
		}}
	case model.AWS_SERVICE_CODE_EC2:
		filters = []types.Filter{{
			Field: aws.String("currentGeneration"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Yes"),
		}, {
			Field: aws.String("capacitystatus"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Used"),
		}, {
			Field: aws.String("marketoption"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("OnDemand"),
		}, {
			Field: aws.String("tenancy"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Shared"),
		}}
	case model.AWS_SERVICE_CODE_EFS:
		filters = []types.Filter{{
			Field: aws.String("locationType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AWS Region"),
		}, {
			Field: aws.String("termType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("OnDemand"),
		}}
	case model.AWS_SERVICE_CODE_ELB:
		filters = []types.Filter{{
			Field: aws.String("locationType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AWS Region"),
		}, {
			Field: aws.String("termType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("OnDemand"),
		}}
	case model.AWS_SERVICE_CODE_RDS:
		filters = []types.Filter{{
			Field: aws.String("currentGeneration"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Yes"),
		}, {
			Field: aws.String("termType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("OnDemand"),
		}}
	case model.AWS_SERVICE_CODE_LAMBDA:
		filters = []types.Filter{{
			Field: aws.String("locationType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AWS Region"),
		}, {
			Field: aws.String("productFamily"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("Serverless"),
		}}
	case model.AWS_SERVICE_CODE_S3:
		filters = []types.Filter{{
			Field: aws.String("locationType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AWS Region"),
		}, {
			Field: aws.String("termType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("onDemand"),
		}}
	}

	// {
	// 	Field: aws.String("location"),
	// 	Type:  types.FilterTypeTermMatch,
	// 	Value: aws.String("Asia Pacific (Seoul)"),
	// }

	// Execute command
	process.OperatePriceCommand(as.Context, svc, as.ServiceCode, filters)
}

func (as AwsService) GetPriceListForTest() error {
	filters := []types.Filter{{
		Field: aws.String("locationType"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("AWS Region"),
	}, {
		Field: aws.String("location"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("US East (N. Virginia)"),
	}, {
		Field: aws.String("termType"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("OnDemand"),
	}}

	return process.OperatePriceCommandForTest(as.Context, svc, as.ServiceCode, filters)
}
