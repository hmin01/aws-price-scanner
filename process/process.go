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

func OperatePriceCommand(ctx context.Context, client *pricing.Client, serviceCode string, filters []types.Filter) {
	cpuCore := runtime.NumCPU()
	// Set channel queue (for raw data and processed data)
	iQueue := make(chan model.RawData, 600)
	oQueue := make(chan interface{}, 600)
	// Set channel queue (for process)
	iProc := make(chan model.ProcessResult, cpuCore)
	oProc := make(chan model.ProcessResult, cpuCore)
	eProc := make(chan model.ProcessResult, 1)

	// Set input parameter
	input := &pricing.GetProductsInput{
		Filters:       filters,
		FormatVersion: aws.String(FORMAT_VERSION),
		MaxResults:    int32(100),
		ServiceCode:   aws.String(serviceCode),
	}

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
		case "AmazonEC2":
			oQueue <- transformPriceDataForEC2(data)
		}
	}
	// Exit
	oProc <- model.ProcessResult{Result: true}
}

func mergePriceData(serviceCode string, oQueue <-chan interface{}, eProc chan<- model.ProcessResult) {
	switch serviceCode {
	case "AmazonEC2":
		mergePriceDataForEC2(oQueue, eProc)
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

func transformPriceDataForEC2(rawData model.RawData) model.InfoForEC2 {
	// Set default object for process
	priceInfo := model.PriceForEC2{
		OperatingSystem: rawData.Product.Attributes["operatingSystem"],
		PreInstalledSw:  rawData.Product.Attributes["preInstalledSw"],
	}
	// Get operation code
	operationCode := rawData.Product.Attributes["operation"]

	// Extract price info
	var respData map[string]interface{}
	for _, value := range rawData.Terms.OnDemand.(map[string]interface{}) {
		respData = value.(map[string]interface{})
		break
	}
	for key, value := range respData {
		if key == "priceDimensions" {
			respData = value.(map[string]interface{})
			break
		}
	}
	for _, value := range respData {
		respData = value.(map[string]interface{})
		break
	}
	for key, value := range respData {
		if key == "pricePerUnit" {
			priceInfo.PricePerUnit = value.(map[string]interface{})
		} else if key == "unit" {
			priceInfo.Unit = value.(string)
		} else if key == "description" {
			priceInfo.Description = value.(string)
		}
	}

	return model.InfoForEC2{
		PriceList: map[string]model.PriceForEC2{
			operationCode: priceInfo,
		},
		Product: model.ProductForEC2{
			InstanceFamily:     rawData.Product.Attributes["instanceFamily"],
			InstanceType:       rawData.Product.Attributes["instanceType"],
			Memory:             rawData.Product.Attributes["memory"],
			NetworkPerformance: rawData.Product.Attributes["networkPerformance"],
			PhysicalProcessor:  rawData.Product.Attributes["physicalProcessor"],
			Vcpu:               rawData.Product.Attributes["vcpu"],
		},
		Region: rawData.Product.Attributes["regionCode"],
		Sku:    rawData.Product.Sku,
	}
}

func mergePriceDataForEC2(oQueue <-chan interface{}, eProc chan<- model.ProcessResult) {
	output := make(map[string]map[string]interface{})
	// Merge data
	for data, ok := <-oQueue; ok; data, ok = <-oQueue {
		// Extract region code and information object
		region := data.(model.InfoForEC2).Region
		it := data.(model.InfoForEC2).Product.InstanceType
		// Merge
		if _, ok := output[region]; !ok {
			output[region] = make(map[string]interface{})
			output[region][it] = map[string]interface{}{
				"priceList": data.(model.InfoForEC2).PriceList,
				"product":   data.(model.InfoForEC2).Product,
				"sku":       data.(model.InfoForEC2).Sku,
			}
		} else if _, ok := output[region][it]; !ok {
			output[region][it] = map[string]interface{}{
				"priceList": data.(model.InfoForEC2).PriceList,
				"product":   data.(model.InfoForEC2).Product,
				"sku":       data.(model.InfoForEC2).Sku,
			}
		} else {
			for key, value := range data.(model.InfoForEC2).PriceList {
				((output[region][it]).(map[string]interface{})["priceList"]).(map[string]model.PriceForEC2)[key] = value
			}
		}
	}
	// Write data
	if err := WriteOutput("ec2.json", output); err != nil {
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
