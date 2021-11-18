package process

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	// AWS
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go/aws"

	// Model
	"aws-price-scanner/model"

	// Module
	"aws-price-scanner/module"
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
		MaxResults:    int32(10),
		ServiceCode:   aws.String(tServiceCode),
	}

	output, err := client.GetProducts(ctx, input)
	if err != nil {
		return err
	}

	for _, data := range output.PriceList {
		fmt.Println(data)
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
	go mergePriceData(serviceCode, oQueue, eProc)

	// Create a paginator
	paginator := pricing.NewGetProductsPaginator(client, input)
	// Process logic
	pCnt := 0
	for {
		output, err := paginator.NextPage(ctx)
		if err != nil {
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
		case <-eProc:
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
		case model.AWS_SERVICE_CODE_EC2:
			oQueue <- transformPriceDataForEC2(data)
		case model.AWS_SERVICE_CODE_EBS:
			oQueue <- transformPriceDataForEBS(data)
		case model.AWS_SERVICE_CODE_LAMBDA:
			oQueue <- transformPriceDataForLambda(data)
		case model.AWS_SERVICE_CODE_RDS:
			oQueue <- transformPriceDataForRDS(data)
		case model.AWS_SERVICE_CODE_S3:
			oQueue <- transformPriceDataForS3(data)
		}
	}
	// Exit
	oProc <- model.ProcessResult{Result: true}
}

func mergePriceData(serviceCode string, oQueue <-chan interface{}, eProc chan<- model.ProcessResult) {
	// Set output file name
	filename := "unknown.json"
	// Merge by service
	switch serviceCode {
	case model.AWS_SERVICE_CODE_EC2:
		filename = "ec2.json"
	case model.AWS_SERVICE_CODE_EBS:
		filename = "ebs.json"
	case model.AWS_SERVICE_CODE_LAMBDA:
		filename = "lambda.json"
	case model.AWS_SERVICE_CODE_RDS:
		filename = "rds.json"
	case model.AWS_SERVICE_CODE_S3:
		filename = "s3.json"
	}

	output := make(map[string]map[string]model.ProcessedData)
	// Merge data
	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
		// Extract region code
		region := data.(model.ProcessedData).Region
		// Extract price key
		key := extractPriceKey(serviceCode, data.(model.ProcessedData).Product)
		// Merge
		if _, ok := output[region]; !ok {
			output[region] = make(map[string]model.ProcessedData)
			output[region][key] = model.ProcessedData{
				OnDemand: data.(model.ProcessedData).OnDemand,
				Product:  data.(model.ProcessedData).Product,
				Sku:      data.(model.ProcessedData).Sku,
			}
		} else if _, ok := output[region][key]; !ok {
			output[region][key] = model.ProcessedData{
				OnDemand: data.(model.ProcessedData).OnDemand,
				Product:  data.(model.ProcessedData).Product,
				Sku:      data.(model.ProcessedData).Sku,
			}
		} else {
			for key, value := range data.(model.ProcessedData).OnDemand {
				output[region][key].OnDemand[key] = value
			}
		}
	}

	// Write data
	if err := WriteOutput(filename, output); err != nil {
		eProc <- model.ProcessResult{
			Result:  false,
			Message: err.Error(),
		}
	} else {
		eProc <- model.ProcessResult{
			Result:  true,
			Message: "Data merger completed",
		}
	}
}

func WriteOutput(filename string, output interface{}) error {
	// Create output file
	file, err := module.CreateOutputFile(filename)
	if err != nil {
		return err
	}

	// Transform to byte array
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}
	// Write data
	file.Write(data)
	file.Close()

	return nil
}

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

func extractPriceKey(serviceCode string, data map[string]string) string {
	switch serviceCode {
	case model.AWS_SERVICE_CODE_EC2:
		return data["instanceType"]
	case model.AWS_SERVICE_CODE_EBS:
		return data["volumeApiName"]
	case model.AWS_SERVICE_CODE_LAMBDA:
		return data["group"]
	case model.AWS_SERVICE_CODE_RDS:
		return data["instanceType"]
	case model.AWS_SERVICE_CODE_S3:
		return data["volumeType"]
	default:
		panic("Invalid service code")
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
		Product: transformDataForInstance(rawData),
		Region:  rawData.Product.Attributes["regionCode"],
		Sku:     rawData.Product.Sku,
	}
}

func transformPriceDataForEBS(rawData model.RawData) model.ProcessedData {
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"storage": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"maxIopsvolume":       rawData.Product.Attributes["maxIopsvolume"],
			"maxThroughputvolume": rawData.Product.Attributes["maxThroughputvolume"],
			"maxVolumeSize":       rawData.Product.Attributes["maxVolumeSize"],
			"storageMedia":        rawData.Product.Attributes["storageMedia"],
			"volumeApiName":       rawData.Product.Attributes["volumeApiName"],
			"volumeType":          rawData.Product.Attributes["volumeType"],
		},
		Region: rawData.Product.Attributes["regionCode"],
		Sku:    rawData.Product.Sku,
	}
}

func transformPriceDataForLambda(rawData model.RawData) model.ProcessedData {
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"storage": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"group":            rawData.Product.Attributes["group"],
			"groupDescription": rawData.Product.Attributes["groupDescription"],
		},
		Region: rawData.Product.Attributes["regionCode"],
		Sku:    rawData.Product.Sku,
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
		Product: transformDataForInstance(rawData),
		Region:  rawData.Product.Attributes["regionCode"],
		Sku:     rawData.Product.Sku,
	}
}

func transformPriceDataForS3(rawData model.RawData) model.ProcessedData {
	return model.ProcessedData{
		OnDemand: map[string][]map[string]interface{}{
			"storage": transformDataForPricePerUnit(rawData.Terms.OnDemand.(map[string]interface{})),
		},
		Product: map[string]string{
			"storageClass": rawData.Product.Attributes["storageClass"],
			"volumeType":   rawData.Product.Attributes["volumeType"],
		},
		Region: rawData.Product.Attributes["regionCode"],
		Sku:    rawData.Product.Sku,
	}
}

// func mergePriceDataForEC2(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) map[string]map[string]interface{} {
// 	output := make(map[string]map[string]interface{})
// 	// Merge data
// 	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
// 		// Extract region code and information object
// 		region := data.(model.InfoForEC2).Region
// 		it := data.(model.InfoForEC2).Product.InstanceType
// 		// Merge
// 		if _, ok := output[region]; !ok {
// 			output[region] = make(map[string]interface{})
// 			output[region][it] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForEC2).OnDemand,
// 				"product":  data.(model.InfoForEC2).Product,
// 				"sku":      data.(model.InfoForEC2).Sku,
// 			}
// 		} else if _, ok := output[region][it]; !ok {
// 			output[region][it] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForEC2).OnDemand,
// 				"product":  data.(model.InfoForEC2).Product,
// 				"sku":      data.(model.InfoForEC2).Sku,
// 			}
// 		} else {
// 			for key, value := range data.(model.InfoForEC2).OnDemand {
// 				((output[region][it]).(map[string]interface{})["onDemand"]).(map[string][]model.OnDemandPriceForEC2)[key] = value
// 			}
// 		}
// 	}
// 	// Return
// 	return output
// }

// func mergePriceDataForEBS(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) map[string]map[string]interface{} {
// 	output := make(map[string]map[string]interface{})
// 	// Merge data
// 	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
// 		// Extract region code and information object
// 		region := data.(model.InfoForEBS).Region
// 		van := data.(model.InfoForEBS).Product.VolumeApiName
// 		// Merge
// 		if _, ok := output[region]; !ok {
// 			output[region] = make(map[string]interface{})
// 			output[region][van] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForEBS).OnDemand,
// 				"product":  data.(model.InfoForEBS).Product,
// 				"sku":      data.(model.InfoForEBS).Sku,
// 			}
// 		} else if _, ok := output[region][van]; !ok {
// 			output[region][van] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForEBS).OnDemand,
// 				"product":  data.(model.InfoForEBS).Product,
// 				"sku":      data.(model.InfoForEBS).Sku,
// 			}
// 		} else {
// 			for key, value := range data.(model.InfoForEBS).OnDemand {
// 				((output[region][van]).(map[string]interface{})["onDemand"]).(map[string][]model.BasicPriceInfo)[key] = value
// 			}
// 		}
// 	}
// 	// Return
// 	return output
// }

// func mergePriceDataForLambda(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) map[string]map[string]interface{} {
// 	output := make(map[string]map[string]interface{})
// 	// Merge data
// 	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
// 		// Extract region code and information object
// 		region := data.(model.InfoForLambda).Region
// 		van := data.(model.InfoForLambda).Product.Group
// 		// Merge
// 		if _, ok := output[region]; !ok {
// 			output[region] = make(map[string]interface{})
// 			output[region][van] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForLambda).OnDemand,
// 				"product":  data.(model.InfoForLambda).Product,
// 				"sku":      data.(model.InfoForLambda).Sku,
// 			}
// 		} else if _, ok := output[region][van]; !ok {
// 			output[region][van] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForLambda).OnDemand,
// 				"product":  data.(model.InfoForLambda).Product,
// 				"sku":      data.(model.InfoForLambda).Sku,
// 			}
// 		} else {
// 			for key, value := range data.(model.InfoForLambda).OnDemand {
// 				((output[region][van]).(map[string]interface{})["onDemand"]).(map[string][]model.BasicPriceInfo)[key] = value
// 			}
// 		}
// 	}
// 	// Return
// 	return output
// }

// func mergePriceDataForRDS(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) map[string]map[string]interface{} {
// 	output := make(map[string]map[string]interface{})
// 	// Merge data
// 	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
// 		// Extract region code and information object
// 		region := data.(model.InfoForRDS).Region
// 		it := data.(model.InfoForRDS).Product.InstanceType
// 		// Merge
// 		if _, ok := output[region]; !ok {
// 			output[region] = make(map[string]interface{})
// 			output[region][it] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForRDS).OnDemand,
// 				"product":  data.(model.InfoForRDS).Product,
// 				"sku":      data.(model.InfoForRDS).Sku,
// 			}
// 		} else if _, ok := output[region][it]; !ok {
// 			output[region][it] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForRDS).OnDemand,
// 				"product":  data.(model.InfoForRDS).Product,
// 				"sku":      data.(model.InfoForRDS).Sku,
// 			}
// 		} else {
// 			for key, value := range data.(model.InfoForRDS).OnDemand {
// 				((output[region][it]).(map[string]interface{})["onDemand"]).(map[string][]model.OnDemandPriceForRDS)[key] = value
// 			}
// 		}
// 	}
// 	// Return
// 	return output
// }

// func mergePriceDataForS3(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) map[string]map[string]interface{} {
// 	output := make(map[string]map[string]interface{})
// 	// Merge data
// 	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
// 		// Extract region code and information object
// 		region := data.(model.InfoForS3).Region
// 		vt := data.(model.InfoForS3).Product.VolumeType
// 		// Merge
// 		if _, ok := output[region]; !ok {
// 			output[region] = make(map[string]interface{})
// 			output[region][vt] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForS3).OnDemand,
// 				"product":  data.(model.InfoForS3).Product,
// 				"sku":      data.(model.InfoForS3).Sku,
// 			}
// 		} else if _, ok := output[region][vt]; !ok {
// 			output[region][vt] = map[string]interface{}{
// 				"onDemand": data.(model.InfoForS3).OnDemand,
// 				"product":  data.(model.InfoForS3).Product,
// 				"sku":      data.(model.InfoForS3).Sku,
// 			}
// 		} else {
// 			for key, value := range data.(model.InfoForS3).OnDemand {
// 				((output[region][vt]).(map[string]interface{})["onDemand"]).(map[string][]model.BasicPriceInfo)[key] = value
// 			}
// 		}
// 	}
// 	// Return
// 	return output
// }