package process

import (
	"strings"

	// Model
	"aws-price-scanner/model"
)

func transformDataForInstance(rawData model.RawData) map[string]string {
	return map[string]string{
		"instanceFamily":    rawData.Product.Attributes["instanceFamily"],
		"instanceType":      rawData.Product.Attributes["instanceType"],
		"memory":            rawData.Product.Attributes["memory"],
		"networkPerformace": rawData.Product.Attributes["networkPerformance"],
		"physicalProcessor": rawData.Product.Attributes["physicalProcessor"],
		"storage":           rawData.Product.Attributes["storage"],
		"vcpu":              rawData.Product.Attributes["vcpu"],
	}
}

func transformDataForPricePerUnit(rawData map[string]interface{}) []map[string]interface{} {
	// Find price dimension
	var respData map[string]interface{}
	for _, value := range rawData {
		respData = value.(map[string]interface{})
		break
	}
	for key, value := range respData {
		if key == "priceDimensions" {
			respData = value.(map[string]interface{})
			break
		}
	}
	// Extract price data
	cnt := 0
	result := make([]map[string]interface{}, len(respData))
	for _, value := range respData {
		obj := make(map[string]interface{})
		data := value.(map[string]interface{})
		for key, value := range data {
			if key == "beginRange" || key == "description" || key == "endRange" || key == "unit" {
				obj[key] = value.(string)
			} else if key == "pricePerUnit" {
				obj["pricePerUnit"] = value
			}
		}
		result[cnt] = obj
		cnt++
	}
	// Return
	return result
}

func transformPriceDataForDynamoDB(rawData model.RawData) model.ProcessedData {
	// Set product type
	productType := "none"
	if rawData.Product.ProductFamily == "Provisioned IOPS" {
		productType = "provisioned"
	} else if rawData.Product.ProductFamily == "Amazon DynamoDB PayPerRequest Throughput" {
		productType = "request"
	} else if rawData.Product.ProductFamily == "Database Storage" {
		productType = "storage"
	}
	// Set service type
	serviceType := "none"
	if productType == "storage" {
		if rawData.Product.Attributes["volumeType"] == "Amazon DynamoDB - Indexed DataStore - IA" {
			serviceType = "Indexed-IA"
		} else if rawData.Product.Attributes["volumeType"] == "Amazon DynamoDB - Indexed DataStore" {
			serviceType = "Indexed"
		} else {
			productType = "none"
		}
	} else {
		if rawData.Product.Attributes["group"] == "DDB-ReadUnits" {
			serviceType = "read"
		} else if rawData.Product.Attributes["group"] == "DDB-WriteUnits" {
			serviceType = "write"
		} else {
			productType = "none"
		}
	}

	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"operation": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"description": rawData.Product.Attributes["groupDescription"],
			"volumeType":  rawData.Product.Attributes["volumeType"],
		},
		ProductType: productType,
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: serviceType,
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForEC2(rawData model.RawData) model.ProcessedData {
	// Get operation code
	operationCode := rawData.Product.Attributes["operation"]
	// Extract price data
	rawOnDemand := transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{}))
	// Set price data
	onDemand := make([]map[string]interface{}, len(rawOnDemand))
	for i, data := range rawOnDemand {
		onDemand[i] = map[string]interface{}{
			"operatingSystem": rawData.Product.Attributes["operatingSystem"],
			"preInstalledSw":  rawData.Product.Attributes["preInstalledSw"],
		}
		for key, value := range data {
			onDemand[i][key] = value
		}
	}
	// Return
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			operationCode: onDemand,
		},
		Product:     transformDataForInstance(rawData),
		ProductType: "instance",
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: rawData.Product.Attributes["instanceType"],
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForEBS(rawData model.RawData) model.ProcessedData {
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"operation": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"maxIopsvolume":       rawData.Product.Attributes["maxIopsvolume"],
			"maxThroughputvolume": rawData.Product.Attributes["maxThroughputvolume"],
			"maxVolumeSize":       rawData.Product.Attributes["maxVolumeSize"],
			"storageMedia":        rawData.Product.Attributes["storageMedia"],
			"volumeType":          rawData.Product.Attributes["volumeType"],
		},
		ProductType: "storage",
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: rawData.Product.Attributes["volumeApiName"],
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForEFS(rawData model.RawData) model.ProcessedData {
	// Set product type
	productType := "none"
	if rawData.Product.ProductFamily == "Storage" {
		productType = "storage"
	} else if rawData.Product.ProductFamily == "Provisioned Throughput" {
		productType = "throughput"
	}
	// Set service type and operation
	var serviceType string
	var operation string
	if productType == "storage" {
		serviceType = strings.ToLower(rawData.Product.Attributes["storageClass"])
		operation = strings.ToLower(rawData.Product.Attributes["operation"])
		if operation == "" {
			operation = "store"
		}
	} else {
		serviceType = strings.ToLower(rawData.Product.Attributes["throughputClass"])
		operation = "operation"
	}

	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			operation: transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"description": rawData.Product.Attributes["groupDescription"],
			"volumeType":  rawData.Product.Attributes["volumeType"],
		},
		ProductType: productType,
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: serviceType,
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForELB(rawData model.RawData) model.ProcessedData {
	// Set service type
	var serviceType string
	if rawData.Product.ProductFamily == "Load Balancer-Gateway" {
		serviceType = "gateway"
	} else if rawData.Product.ProductFamily == "Load Balancer-Application" {
		serviceType = "application"
	} else if rawData.Product.ProductFamily == "Load Balancer-Network" {
		serviceType = "network"
	} else {
		serviceType = "classic"
	}
	// Set onDemand key
	onDemandKey := "none"
	if serviceType == "classic" {
		if strings.Contains(rawData.Product.Attributes["usagetype"], "LoadBalancerUsage") {
			onDemandKey = "usage"
		} else if strings.Contains(rawData.Product.Attributes["usagetype"], "DataProcessing-Bytes") {
			onDemandKey = "processing"
		} else {
			serviceType = "none"
		}
	} else {
		if strings.Contains(rawData.Product.Attributes["usagetype"], "LoadBalancerUsage") {
			onDemandKey = "usage"
		} else if strings.Contains(rawData.Product.Attributes["usagetype"], "LCUUsage") {
			onDemandKey = "lcu"
		} else {
			serviceType = "none"
		}
	}

	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			onDemandKey: transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"description": rawData.Product.Attributes["groupDescription"],
		},
		ProductType: "loadBalancer",
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: serviceType,
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForLambda(rawData model.RawData) model.ProcessedData {
	// Set product type
	var productType string
	if strings.Contains(rawData.Product.Attributes["group"], "Provisioned") {
		productType = "provisioned"
	} else if strings.Contains(rawData.Product.Attributes["group"], "Edge") {
		productType = "edge"
	} else {
		productType = "usual"
	}
	// Set onDemand key
	var onDemandKey string
	if strings.Contains(rawData.Product.Attributes["group"], "Duration") {
		if strings.Contains(rawData.Product.Attributes["group"], "ARM") {
			onDemandKey = "duration-arm"
		} else {
			onDemandKey = "duration"
		}
	} else {
		if strings.Contains(rawData.Product.Attributes["group"], "ARM") {
			onDemandKey = "requests-arm"
		} else {
			onDemandKey = "requests"
		}
	}

	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			onDemandKey: transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		ProductType: productType,
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: "function",
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForRDS(rawData model.RawData) model.ProcessedData {
	// Get operation code
	operationCode := rawData.Product.Attributes["operation"]
	// Extract price data
	rawOnDemand := transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{}))
	// Set price data
	onDemand := make([]map[string]interface{}, len(rawOnDemand))
	for i, data := range rawOnDemand {
		onDemand[i] = map[string]interface{}{
			"deploymentOption": rawData.Product.Attributes["deploymentOption"],
			"databaseEdition":  rawData.Product.Attributes["databaseEdition"],
			"databaseEngine":   rawData.Product.Attributes["databaseEngine"],
		}
		for key, value := range data {
			onDemand[i][key] = value
		}
	}
	// Return
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			operationCode: onDemand,
		},
		Product:     transformDataForInstance(rawData),
		ProductType: "instance",
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: rawData.Product.Attributes["instanceType"],
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

func transformPriceDataForS3(rawData model.RawData) model.ProcessedData {
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"operation": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"storageClass": strings.ToLower(rawData.Product.Attributes["storageClass"]),
			"volumeType":   rawData.Product.Attributes["volumeType"],
		},
		ProductType: "storage",
		Region:      rawData.Product.Attributes["regionCode"],
		Sku:         rawData.Product.Sku,
		ServiceType: strings.ToLower(rawData.Product.Attributes["volumeType"]),
		UsageType:   rawData.Product.Attributes["usagetype"],
	}
}

// func transformPriceDataForVpcEndpoint(rawData model.RawData) model.ProcessedData {
// 	reEpType := regexp.MustCompile("^\\S+-VpcEndpoint-")
// 	reType := regexp.MustCompile("^GWLBE")
// 	reTarget := regexp.MustCompile("Bytes")
// 	// Process usagetype
// 	usageType := reEpType.ReplaceAllString(rawData.Product.Attributes["usagetype"], "")
// 	// Find endpoint type (interface or gateway)
// 	var epType string
// 	if reType.MatchString(usageType) {
// 		epType = "gateway"
// 	} else {
// 		epType = "interface"
// 	}
// 	// Find price target
// 	var priceTarget string
// 	if reTarget.MatchString(usageType) {
// 		priceTarget = "process"
// 	} else {
// 		priceTarget = "usage"
// 	}

// 	return model.ProcessedData{
// 		OnDemand: map[string][]map[string]interface{}{
// 			priceTarget: transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
// 		},
// 		Product: map[string]string{
// 			"type":  epType,
// 			"group": rawData.Product.Attributes["group"],
// 		},
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }
