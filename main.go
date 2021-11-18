package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"

	// Custom aws module
	"aws-price-scanner/aws"
	"aws-price-scanner/model"
	// Model
)

var testServiceCode = "AWSLambda"

func init() {
	// Configure an AWS SDK
	if err := aws.Configure(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.TODO()

	// Create flag to use command line
	serviceCode := command()
	// Process command
	if serviceCode == "Unknown" {
		flag.Usage()
	} else if serviceCode == "" {
		// Get a list of service code
		list, err := aws.ServiceCodeList(ctx)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("*-- Service code list ---*")
			for _, elem := range list {
				fmt.Println(elem)
			}
		}
	} else {
		srv := aws.NewService(ctx, serviceCode)
		srv.GetPriceList()
	}

	// Test1(ctx)

	// Test2(ctx)
}

func command() string {
	// Create usage Description
	var desc bytes.Buffer
	desc.WriteString("AWS service code\nSupport a list of service code: ")
	for i, code := range model.AWS_SERVICE_CODE_LIST {
		desc.WriteString(code)
		if i < len(model.AWS_SERVICE_CODE_LIST)-1 {
			desc.WriteString(", ")
		}
	}
	// Create flag
	srvFlag := flag.String("srv", "", desc.String())
	flag.Parse()

	// Check flag
	if flag.NFlag() == 0 {
		return ""
	} else {
		match := false
		for _, code := range model.AWS_SERVICE_CODE_LIST {
			if code == *srvFlag {
				match = true
				break
			}
		}
		// Return
		if match {
			return *srvFlag
		} else {
			return "Unknown"
		}
	}
}

func Test1(ctx context.Context) {
	// Create service
	srv := aws.NewService(ctx, testServiceCode)
	// Get attributes for service
	attributes, err := srv.GetAttributes()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("*--- Attributes ---*")
		fmt.Println(attributes)
		fmt.Println()
	}
	// Get attribute values for service
	for _, attribute := range attributes {
		values, err := srv.GetAttributeValues(attribute)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("*--- " + attribute + " ---*")
			fmt.Println(values)
			fmt.Println()
		}
	}
}

func Test2(ctx context.Context) {
	// Create service
	srv := aws.NewService(ctx, testServiceCode)
	// Get price list (for test)
	if err := srv.GetPriceListForTest(); err != nil {
		log.Fatal(err)
	}
}
