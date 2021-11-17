package model

const (
	AWS_SERVICE_CODE_EBS string = "AmazonEBS"
	AWS_SERVICE_CODE_EC2 string = "AmazonEC2"
	AWS_SERVICE_CODE_RDS string = "AmazonRDS"
)

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

type ProductForInstance struct {
	DedicatedEbsThroughput string `json:"dedicatedEbsThroughput,omitempty"`
	InstanceFamily         string `json:"instanceFamily"`
	InstanceType           string `json:"instanceType"`
	Memory                 string `json:"memory"`
	NetworkPerformance     string `json:"networkPerformance"`
	PhysicalProcessor      string `json:"physicalProcessor"`
	Storage                string `json:"storage"`
	Vcpu                   string `json:"vcpu"`
}
type BasicPriceInfo struct {
	BeginRange   string                 `json:"beginRange"`
	Description  string                 `json:"description"`
	EndRange     string                 `json:"endRange"`
	PricePerUnit map[string]interface{} `json:"pricePerUnit"`
	Unit         string                 `json:"unit"`
}

/* For EC2 */
type InfoForEC2 struct {
	OnDemand map[string]OnDemandPriceForEC2 `json:"onDemand"`
	Product  ProductForInstance             `json:"product"`
	Region   string                         `json:"region"`
	Reserved map[string]ReservedPriceForEC2 `json:"reserved"`
	Sku      string                         `json:"sku"`
}
type OnDemandPriceForEC2 struct {
	BasicPriceInfo
	OperatingSystem string `json:"operatingSystem"`
	PreInstalledSw  string `json:"preInstalledSw"`
}
type ReservedPriceForEC2 struct {
	BasicPriceInfo  `json:"price"`
	OperatingSystem string `json:"operatingSystem"`
	PreInstalledSw  string `json:"preInstalledSw"`
	Term            struct {
		LeaseContractLength string `json:"leaseContractLength"`
		OfferingClass       string `json:"offeringClass"`
		PurchaseOption      string `json:"purchaseOption"`
	} `json:"term"`
}

/* For EBS */
type InfoForEBS struct {
	OnDemand map[string]OnDemandPriceForEBS `json:"onDemand"`
	Product  ProductForEBS                  `json:"product"`
	Region   string                         `json:"region"`
	Sku      string                         `json:"sku"`
}
type ProductForEBS struct {
	MaxIopsvolume       string `json:"maxIopsvolume"`
	MaxThroughputvolume string `json:"maxThroughputvolume"`
	MaxVolumeSize       string `json:"maxVolumeSize"`
	StorageMedia        string `json:"storageMedia"`
	VolumeApiName       string `json:"volumeApiName"`
	VolumeType          string `json:"volumeType"`
}
type OnDemandPriceForEBS struct {
	BasicPriceInfo `json:"price"`
}

/* For RDS */
type InfoForRDS struct {
	OnDemand map[string]OnDemandPriceForRDS `json:"onDemand"`
	Product  ProductForInstance             `json:"product"`
	Region   string                         `json:"region"`
	Reserved map[string]ReservedPriceForRDS `json:"reserved"`
	Sku      string                         `json:"sku"`
}
type OnDemandPriceForRDS struct {
	BasicPriceInfo   `json:"price"`
	DeploymentOption string `json:"deploymentOption"`
	DatabaseEdition  string `json:"databaseEdition"`
	DatabaseEngine   string `json:"databaseEngine"`
}
type ReservedPriceForRDS struct {
	BasicPriceInfo
	DeploymentOption string `json:"deploymentOption"`
	DatabaseEdition  string `json:"databaseEdition"`
	DatabaseEngine   string `json:"databaseEngine"`
	Term             struct {
		LeaseContractLength string `json:"leaseContractLength"`
		OfferingClass       string `json:"offeringClass"`
		PurchaseOption      string `json:"purchaseOption"`
	} `json:"term"`
}
