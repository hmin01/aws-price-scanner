package model

const (
	AWS_SERVICE_CODE_EBS    = "AmazonEBS"
	AWS_SERVICE_CODE_EC2    = "AmazonEC2"
	AWS_SERVICE_CODE_LAMBDA = "AWSLambda"
	AWS_SERVICE_CODE_RDS    = "AmazonRDS"
	AWS_SERVICE_CODE_S3     = "AmazonS3"
)

var AWS_SERVICE_CODE_LIST = []string{"AmazonEBS", "AmazonEC2", "AWSLambda", "AmazonRDS", "AmazonS3"}

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
	OnDemand map[string][]map[string]interface{} `json:"onDemand"`
	Product  map[string]string                   `json:"product"`
	Region   string                              `json:"region,omitempty"`
	Sku      string                              `json:"sku"`
}

// type ProductForInstance struct {
// 	DedicatedEbsThroughput string `json:"dedicatedEbsThroughput,omitempty"`
// 	InstanceFamily         string `json:"instanceFamily"`
// 	InstanceType           string `json:"instanceType"`
// 	Memory                 string `json:"memory"`
// 	NetworkPerformance     string `json:"networkPerformance"`
// 	PhysicalProcessor      string `json:"physicalProcessor"`
// 	Storage                string `json:"storage"`
// 	Vcpu                   string `json:"vcpu"`
// }
// type BasicPriceInfo struct {
// 	BeginRange   string                 `json:"beginRange"`
// 	Description  string                 `json:"description"`
// 	EndRange     string                 `json:"endRange"`
// 	PricePerUnit map[string]interface{} `json:"pricePerUnit"`
// 	Unit         string                 `json:"unit"`
// }

// /* For VPC */
// // type InfoForVPC struct {
// // 	OnDemand map[string]
// // }
// // type OnDemandPriceForVPC struct {
// // 	BasicPriceInfo
// // }

// /* For EC2 */
// type InfoForEC2 struct {
// 	OnDemand map[string][]OnDemandPriceForEC2 `json:"onDemand"`
// 	Product  ProductForInstance               `json:"product"`
// 	Region   string                           `json:"region"`
// 	Reserved map[string][]ReservedPriceForEC2 `json:"reserved"`
// 	Sku      string                           `json:"sku"`
// }
// type OnDemandPriceForEC2 struct {
// 	BasicPriceInfo
// 	OperatingSystem string `json:"operatingSystem"`
// 	PreInstalledSw  string `json:"preInstalledSw"`
// }
// type ReservedPriceForEC2 struct {
// 	BasicPriceInfo  `json:"price"`
// 	OperatingSystem string `json:"operatingSystem"`
// 	PreInstalledSw  string `json:"preInstalledSw"`
// 	Term            struct {
// 		LeaseContractLength string `json:"leaseContractLength"`
// 		OfferingClass       string `json:"offeringClass"`
// 		PurchaseOption      string `json:"purchaseOption"`
// 	} `json:"term"`
// }

// /* For EBS */
// type InfoForEBS struct {
// 	OnDemand map[string][]BasicPriceInfo `json:"onDemand"`
// 	Product  ProductForEBS               `json:"product"`
// 	Region   string                      `json:"region"`
// 	Sku      string                      `json:"sku"`
// }
// type ProductForEBS struct {
// 	MaxIopsvolume       string `json:"maxIopsvolume"`
// 	MaxThroughputvolume string `json:"maxThroughputvolume"`
// 	MaxVolumeSize       string `json:"maxVolumeSize"`
// 	StorageMedia        string `json:"storageMedia"`
// 	VolumeApiName       string `json:"volumeApiName"`
// 	VolumeType          string `json:"volumeType"`
// }

// /* For RDS */
// type InfoForRDS struct {
// 	OnDemand map[string][]OnDemandPriceForRDS `json:"onDemand"`
// 	Product  ProductForInstance               `json:"product"`
// 	Region   string                           `json:"region"`
// 	Reserved map[string][]ReservedPriceForRDS `json:"reserved"`
// 	Sku      string                           `json:"sku"`
// }
// type OnDemandPriceForRDS struct {
// 	BasicPriceInfo   `json:"price"`
// 	DeploymentOption string `json:"deploymentOption"`
// 	DatabaseEdition  string `json:"databaseEdition"`
// 	DatabaseEngine   string `json:"databaseEngine"`
// }
// type ReservedPriceForRDS struct {
// 	BasicPriceInfo
// 	DeploymentOption string `json:"deploymentOption"`
// 	DatabaseEdition  string `json:"databaseEdition"`
// 	DatabaseEngine   string `json:"databaseEngine"`
// 	Term             struct {
// 		LeaseContractLength string `json:"leaseContractLength"`
// 		OfferingClass       string `json:"offeringClass"`
// 		PurchaseOption      string `json:"purchaseOption"`
// 	} `json:"term"`
// }

// /* For Lambda */
// type InfoForLambda struct {
// 	OnDemand map[string][]BasicPriceInfo `json:"onDemand"`
// 	Product  ProductForLambda            `json:"product"`
// 	Region   string                      `json:"region"`
// 	Sku      string                      `json:"sku"`
// }
// type ProductForLambda struct {
// 	Group            string `json:"group"`
// 	GroupDescription string `json:"groupDescription"`
// }

// /* For S3 */
// type InfoForS3 struct {
// 	OnDemand map[string][]BasicPriceInfo `json:"onDemand"`
// 	Product  ProductForS3                `json:"product"`
// 	Region   string                      `json:"region"`
// 	Sku      string                      `json:"sku"`
// }
// type ProductForS3 struct {
// 	StorageClass string `json:"storageClass"`
// 	VolumeType   string `json:"volumeType"`
// }
