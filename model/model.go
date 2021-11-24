package model

const (
	AWS_SERVICE_CODE_EBS          = "AmazonEBS"
	AWS_SERVICE_CODE_EC2          = "AmazonEC2"
	AWS_SERVICE_CODE_LAMBDA       = "AWSLambda"
	AWS_SERVICE_CODE_RDS          = "AmazonRDS"
	AWS_SERVICE_CODE_S3           = "AmazonS3"
	AWS_SERVICE_CODE_VPC_ENDPOINT = "AmazonVpc"
)

var AWS_SERVICE_CODE_LIST = []string{"AmazonEBS", "AmazonEC2", "AWSLambda", "AmazonRDS", "AmazonS3", "AmazonVpc"}

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
	OnDemand  map[string][]map[string]interface{} `json:"onDemand"`
	Product   map[string]string                   `json:"product"`
	Region    string                              `json:"region,omitempty"`
	Sku       string                              `json:"sku"`
	UsageType string                              `json:"usageType"`
}
