package model

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

/* For EC2 */
type InfoForEC2 struct {
	PriceList map[string]PriceForEC2 `json:"priceList"`
	Product   ProductForEC2          `json:"product"`
	Region    string                 `json:"region"`
	Sku       string                 `json:"sku"`
}
type ProductForEC2 struct {
	InstanceFamily     string `json:"instanceFamily"`
	InstanceType       string `json:"instanceType"`
	Memory             string `json:"memory"`
	NetworkPerformance string `json:"networkPerformance"`
	PhysicalProcessor  string `json:"physicalProcessor"`
	Vcpu               string `json:"vcpu"`
}
type PriceForEC2 struct {
	Description     string                 `json:"description"`
	OperatingSystem string                 `json:"operatingSystem"`
	PreInstalledSw  string                 `json:"preInstalledSw"`
	PricePerUnit    map[string]interface{} `json:"pricePerUnit"`
	Unit            string                 `json:"unit"`
}

/* For RDS */
type InfoForRDS struct {
}
type ProductForRDS struct {
}
type PriceForRDS struct {
}
