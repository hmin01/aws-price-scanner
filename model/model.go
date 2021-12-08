package model

const (
	AWS_SERVICE_CODE_DYNAMODB     = "AmazonDynamoDB"
	AWS_SERVICE_CODE_EBS          = "AmazonEBS"
	AWS_SERVICE_CODE_EC2          = "AmazonEC2"
	AWS_SERVICE_CODE_LAMBDA       = "AWSLambda"
	AWS_SERVICE_CODE_RDS          = "AmazonRDS"
	AWS_SERVICE_CODE_S3           = "AmazonS3"
	AWS_SERVICE_CODE_VPC_ENDPOINT = "AmazonVpc"

	CODE_SUCCES                 = 0
	CODE_ERROR_INVAILD_ARGUMENT = 100
	CODE_ERROR_INVALID_S3       = 101
	CODE_ERROR_PROCESS_FAIL     = 104
)

var AWS_SERVICE_CODE_LIST = []string{"AmazonDynamoDB", "AmazonEBS", "AmazonEC2", "AWSLambda", "AmazonRDS", "AmazonS3", "AmazonVpc"}

type ProcessResult struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

type RawData struct {
	Product struct {
		Attributes map[string]string `json:"attributes"`
		Sku        string            `json:"sku"`
	} `json:"product"`
	Terms struct {
		OnDemand interface{} `json:"OnDemand"`
	} `json:"terms"`
	Version string `json:"version"`
}

type ProcessedData struct {
	DistKey   string                              `json:"DistKey"`
	OnDemand  map[string][]map[string]interface{} `json:"onDemand"`
	Product   map[string]string                   `json:"product"`
	Region    string                              `json:"region,omitempty"`
	Sku       string                              `json:"sku"`
	UsageType string                              `json:"usageType"`
}
