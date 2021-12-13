package process

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	// AWS
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"

	// Model
	"aws-price-scanner/model"
	// Custom S3
	"aws-price-scanner/aws/s3"
)

const FORMAT_VERSION = "aws_v1"

func OperatePriceCommandForTest(ctx context.Context, client *pricing.Client, serviceCode string, filters []types.Filter) error {
	// If service code is EBS
	tServiceCode := serviceCode
	if tServiceCode == model.AWS_SERVICE_CODE_EBS {
		tServiceCode = model.AWS_SERVICE_CODE_EC2
	}
	// Set input parameter
	input := &pricing.GetProductsInput{
		Filters:       filters,
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(20),
		ServiceCode:   aws.String(tServiceCode),
	}

	output, err := client.GetProducts(ctx, input)
	if err != nil {
		return err
	}

	for _, data := range output.PriceList {
		fmt.Println(data)
		fmt.Println()
	}
	return nil
}

func OperatePriceCommand(ctx context.Context, client *pricing.Client, serviceCode string, filters []types.Filter) {
	cpuCore := runtime.NumCPU()
	// Set channel queue (for raw data and processed data)
	iQueue := make(chan model.RawData, 600)
	oQueue := make(chan interface{}, 600)
	// Set channel queue (for process)
	iProc := make(chan model.ProcessResult, cpuCore)
	oProc := make(chan model.ProcessResult, cpuCore)
	eProc := make(chan model.ProcessResult, 1)

	// If service code is EBS
	tServiceCode := serviceCode
	if tServiceCode == model.AWS_SERVICE_CODE_EBS {
		tServiceCode = model.AWS_SERVICE_CODE_EC2
	}
	// Set input parameter
	input := &pricing.GetProductsInput{
		Filters:       filters,
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(100),
		ServiceCode:   aws.String(tServiceCode),
	}

	fmt.Println("Configure complete")
	fmt.Println("Processing...")

	// Execute process (extract and transform data, merge transformed data)
	for i := 0; i < cpuCore; i++ {
		go transformPriceData(serviceCode, iQueue, oQueue, oProc)
	}
	go mergePriceData(ctx, serviceCode, oQueue, eProc)

	// Create a paginator
	paginator := pricing.NewGetProductsPaginator(client, input)
	// Process logic
	pCnt := 0
	for {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Println("[ERROR] " + err.Error())
			return
		}
		go extractPriceData(output, iQueue, iProc)
		pCnt++
		// Escape
		if !paginator.HasMorePages() {
			break
		}
	}

	// Exit logic
	iCompleted := 0
	oCompleted := 0
	for {
		select {
		case <-iProc:
			iCompleted++
			if iCompleted >= pCnt {
				close(iQueue)
				// Print message
				fmt.Println("Request data complete.")
			}
		case <-oProc:
			oCompleted++
			if oCompleted >= runtime.NumCPU() {
				close(oQueue)
				// Print message
				fmt.Println("Transform data completed.")
			}
		case result := <-eProc:
			if result.Result {
				fmt.Println(result.Message)
				os.Exit(model.CODE_SUCCES)
			} else {
				fmt.Println("[ERROR] " + result.Message)
				os.Exit(model.CODE_ERROR_PROCESS_FAIL)
			}
			return
		}
	}
}

func extractPriceData(output *pricing.GetProductsOutput, iQueue chan<- model.RawData, iProc chan<- model.ProcessResult) {
	for _, data := range output.PriceList {
		// Transform
		var rawData model.RawData
		if err := json.Unmarshal([]byte(data), &rawData); err != nil {
			iProc <- model.ProcessResult{
				Result:  false,
				Message: err.Error(),
			}
		}
		// Push data
		iQueue <- rawData
	}
	// Exit
	iProc <- model.ProcessResult{Result: true}
}

func transformPriceData(serviceCode string, iQueue <-chan model.RawData, oQueue chan<- interface{}, oProc chan<- model.ProcessResult) {
	for data, ok := <-iQueue; ok; data, ok = <-iQueue {
		switch serviceCode {
		case model.AWS_SERVICE_CODE_DYNAMODB:
			oQueue <- transformPriceDataForDynamoDB(data)
		case model.AWS_SERVICE_CODE_EBS:
			oQueue <- transformPriceDataForEBS(data)
		case model.AWS_SERVICE_CODE_EC2:
			oQueue <- transformPriceDataForEC2(data)
		case model.AWS_SERVICE_CODE_ECS:
			oQueue <- transformPriceDataForECS(data)
		case model.AWS_SERVICE_CODE_EFS:
			oQueue <- transformPriceDataForEFS(data)
		case model.AWS_SERVICE_CODE_ELB:
			oQueue <- transformPriceDataForELB(data)
		case model.AWS_SERVICE_CODE_LAMBDA:
			oQueue <- transformPriceDataForLambda(data)
		case model.AWS_SERVICE_CODE_RDS:
			oQueue <- transformPriceDataForRDS(data)
		case model.AWS_SERVICE_CODE_S3:
			oQueue <- transformPriceDataForS3(data)
		case model.AWS_SERVICE_CODE_VPC:
			oQueue <- transformPriceDataForVPC(data)
		}
	}
	// Exit
	oProc <- model.ProcessResult{Result: true}
}

func mergePriceData(ctx context.Context, serviceCode string, oQueue <-chan interface{}, eProc chan<- model.ProcessResult) {
	// // Set output file name
	filename := serviceCode + ".json"

	// Merge data
	output := make(map[string]map[string]map[string]map[string]interface{})
	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
		// Check distinguish key (if it is "none", not processing)
		productType := data.(model.ProcessedData).ProductType
		if productType == "none" {
			continue
		}

		// Extract region code and service type
		region := data.(model.ProcessedData).Region
		serviceType := data.(model.ProcessedData).ServiceType
		// Merge
		if _, ok := output[region]; !ok {
			output[region] = make(map[string]map[string]map[string]interface{})
			output[region][productType] = map[string]map[string]interface{}{
				serviceType: make(map[string]interface{}),
			}
			output[region][productType][serviceType] = map[string]interface{}{
				"onDemand": make(map[string][]map[string]interface{}),
				"product":  data.(model.ProcessedData).Product,
				"sku":      data.(model.ProcessedData).Sku,
			}
			for key, value := range data.(model.ProcessedData).OnDemand {
				(output[region][productType][serviceType]["onDemand"]).(map[string][]map[string]interface{})[key] = value
			}
		} else if _, ok := output[region][productType]; !ok {
			output[region][productType] = map[string]map[string]interface{}{
				serviceType: make(map[string]interface{}),
			}
			output[region][productType][serviceType] = map[string]interface{}{
				"onDemand": make(map[string][]map[string]interface{}),
				"product":  data.(model.ProcessedData).Product,
				"sku":      data.(model.ProcessedData).Sku,
			}
			for key, value := range data.(model.ProcessedData).OnDemand {
				(output[region][productType][serviceType]["onDemand"]).(map[string][]map[string]interface{})[key] = value
			}
		} else if _, ok := output[region][productType][serviceType]; !ok {
			output[region][productType][serviceType] = map[string]interface{}{
				"onDemand": make(map[string][]map[string]interface{}),
				"product":  data.(model.ProcessedData).Product,
				"sku":      data.(model.ProcessedData).Sku,
			}
			for key, value := range data.(model.ProcessedData).OnDemand {
				(output[region][productType][serviceType]["onDemand"]).(map[string][]map[string]interface{})[key] = value
			}
		} else {
			for key, value := range data.(model.ProcessedData).OnDemand {
				(output[region][productType][serviceType]["onDemand"]).(map[string][]map[string]interface{})[key] = value
			}
		}
	}

	if err := s3.UploadOutput(ctx, filename, output); err != nil {
		eProc <- model.ProcessResult{
			Result:  false,
			Message: err.Error(),
		}
	} else {
		eProc <- model.ProcessResult{
			Result:  true,
			Message: "Process completed",
		}
	}

	// // Write data
	// if err := WriteOutput(filename, output); err != nil {
	// 	eProc <- model.ProcessResult{
	// 		Result:  false,
	// 		Message: err.Error(),
	// 	}
	// } else {
	// 	eProc <- model.ProcessResult{
	// 		Result:  true,
	// 		Message: "Data merger completed",
	// 	}
	// }
}

// func WriteOutput(filename string, output interface{}) error {
// 	// Create output file
// 	file, err := module.CreateOutputFile(filename)
// 	if err != nil {
// 		return err
// 	}

// 	// Transform to byte array
// 	data, err := json.Marshal(output)
// 	if err != nil {
// 		return err
// 	}
// 	// Write data
// 	file.Write(data)
// 	file.Close()

// 	return nil
// }

// func transformDataForInstance(rawData model.RawData) map[string]string {
// 	return map[string]string{
// 		"instanceFamily":    rawData.Product.Attributes["instanceFamily"],
// 		"instanceType":      rawData.Product.Attributes["instanceType"],
// 		"memory":            rawData.Product.Attributes["memory"],
// 		"networkPerformace": rawData.Product.Attributes["networkPerformance"],
// 		"physicalProcessor": rawData.Product.Attributes["physicalProcessor"],
// 		"storage":           rawData.Product.Attributes["storage"],
// 		"vcpu":              rawData.Product.Attributes["vcpu"],
// 	}
// }

// func transformDataForPricePerUnit(rawData map[string]interface{}) []map[string]interface{} {
// 	// Find price dimension
// 	var respData map[string]interface{}
// 	for _, value := range rawData {
// 		respData = value.(map[string]interface{})
// 		break
// 	}
// 	for key, value := range respData {
// 		if key == "priceDimensions" {
// 			respData = value.(map[string]interface{})
// 			break
// 		}
// 	}
// 	// Extract price data
// 	cnt := 0
// 	result := make([]map[string]interface{}, len(respData))
// 	for _, value := range respData {
// 		obj := make(map[string]interface{})
// 		data := value.(map[string]interface{})
// 		for key, value := range data {
// 			if key == "beginRange" || key == "description" || key == "endRange" || key == "unit" {
// 				obj[key] = value.(string)
// 			} else if key == "pricePerUnit" {
// 				obj["pricePerUnit"] = value
// 			}
// 		}
// 		result[cnt] = obj
// 		cnt++
// 	}
// 	// Return
// 	return result
// }

// func extractPriceKey(serviceCode string, data map[string]string) string {
// 	switch serviceCode {
// 	case model.AWS_SERVICE_CODE_EC2:
// 		return data["instanceType"]
// 	case model.AWS_SERVICE_CODE_EBS:
// 		return data["volumeApiName"]
// 	case model.AWS_SERVICE_CODE_LAMBDA:
// 		return data["group"]
// 	case model.AWS_SERVICE_CODE_RDS:
// 		return data["instanceType"]
// 	case model.AWS_SERVICE_CODE_S3:
// 		return data["volumeType"]
// 	case model.AWS_SERVICE_CODE_VPC_ENDPOINT:
// 		return data["type"]
// 	default:
// 		panic("Invalid service code")
// 	}
// }

// func transformPriceDataForEC2(rawData model.RawData) model.ProcessedData {
// 	// Get operation code
// 	operationCode := rawData.Product.Attributes["operation"]
// 	// Extract price data
// 	rawOnDemand := transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{}))
// 	// Set price data
// 	onDemand := make([]map[string]interface{}, len(rawOnDemand))
// 	for i, data := range rawOnDemand {
// 		onDemand[i] = map[string]interface{}{
// 			"operatingSystem": rawData.Product.Attributes["operatingSystem"],
// 			"preInstalledSw":  rawData.Product.Attributes["preInstalledSw"],
// 		}
// 		for key, value := range data {
// 			onDemand[i][key] = value
// 		}
// 	}
// 	// Return
// 	return model.ProcessedData{
// 		DistKey: rawData.Product.Attributes["instanceType"],
// 		OnDemand: map[string][]map[string]interface{}{
// 			operationCode: onDemand,
// 		},
// 		Product:   transformDataForInstance(rawData),
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }

// func transformPriceDataForEBS(rawData model.RawData) model.ProcessedData {
// 	return model.ProcessedData{
// 		DistKey: rawData.Product.Attributes["volumeApiName"],
// 		OnDemand: map[string][]map[string]interface{}{
// 			"storage": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
// 		},
// 		Product: map[string]string{
// 			"maxIopsvolume":       rawData.Product.Attributes["maxIopsvolume"],
// 			"maxThroughputvolume": rawData.Product.Attributes["maxThroughputvolume"],
// 			"maxVolumeSize":       rawData.Product.Attributes["maxVolumeSize"],
// 			"storageMedia":        rawData.Product.Attributes["storageMedia"],
// 			"volumeType":          rawData.Product.Attributes["volumeType"],
// 		},
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }

// func transformPriceDataForLambda(rawData model.RawData) model.ProcessedData {
// 	// Set distinguished key
// 	var distKey string
// 	if strings.Contains(rawData.Product.Attributes["group"], "Provisioned") {
// 		distKey = "provisioned"
// 	} else if strings.Contains(rawData.Product.Attributes["group"], "Edge") {
// 		distKey = "edge"
// 	} else {
// 		distKey = "usual"
// 	}
// 	// Set onDemand key
// 	var onDemandKey string
// 	if strings.Contains(rawData.Product.Attributes["group"], "Duration") {
// 		if strings.Contains(rawData.Product.Attributes["group"], "ARM") {
// 			onDemandKey = "duration-arm"
// 		} else {
// 			onDemandKey = "duration"
// 		}
// 	} else {
// 		if strings.Contains(rawData.Product.Attributes["group"], "ARM") {
// 			onDemandKey = "requests-arm"
// 		} else {
// 			onDemandKey = "requests"
// 		}
// 	}

// 	return model.ProcessedData{
// 		DistKey: distKey,
// 		OnDemand: map[string][]map[string]interface{}{
// 			onDemandKey: transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
// 		},
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }

// func transformPriceDataForRDS(rawData model.RawData) model.ProcessedData {
// 	// Get operation code
// 	operationCode := rawData.Product.Attributes["operation"]
// 	// Extract price data
// 	rawOnDemand := transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{}))
// 	// Set price data
// 	onDemand := make([]map[string]interface{}, len(rawOnDemand))
// 	for i, data := range rawOnDemand {
// 		onDemand[i] = map[string]interface{}{
// 			"deploymentOption": rawData.Product.Attributes["deploymentOption"],
// 			"databaseEdition":  rawData.Product.Attributes["databaseEdition"],
// 			"databaseEngine":   rawData.Product.Attributes["databaseEngine"],
// 		}
// 		for key, value := range data {
// 			onDemand[i][key] = value
// 		}
// 	}
// 	// Return
// 	return model.ProcessedData{
// 		DistKey: rawData.Product.Attributes["instanceType"],
// 		OnDemand: map[string][]map[string]interface{}{
// 			operationCode: onDemand,
// 		},
// 		Product:   transformDataForInstance(rawData),
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }

// func transformPriceDataForS3(rawData model.RawData) model.ProcessedData {
// 	return model.ProcessedData{
// 		DistKey: strings.ToLower(rawData.Product.Attributes["storageClass"]),
// 		OnDemand: map[string][]map[string]interface{}{
// 			"storage": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
// 		},
// 		Product: map[string]string{
// 			"storageClass": strings.ToLower(rawData.Product.Attributes["storageClass"]),
// 			"volumeType":   rawData.Product.Attributes["volumeType"],
// 		},
// 		Region:    rawData.Product.Attributes["regionCode"],
// 		Sku:       rawData.Product.Sku,
// 		UsageType: rawData.Product.Attributes["usagetype"],
// 	}
// }

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
