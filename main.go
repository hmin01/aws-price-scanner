package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	// Custom aws module
	"aws-price-scanner/aws"
	"aws-price-scanner/model"
	// Model
)

func init() {
	// Configure an AWS SDK
	if err := aws.Configure(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.TODO()

	// // Create flag to use command line
	// serviceCode := command()
	// // Process command
	// if serviceCode == "Unknown" {
	// 	flag.Usage()
	// } else if serviceCode == "" {
	// 	// Get a list of service code
	// 	list, err := aws.ServiceCodeList(ctx)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	} else {
	// 		fmt.Println("*-- Service code list ---*")
	// 		for _, elem := range list {
	// 			fmt.Println(elem)
	// 		}
	// 	}
	// } else {
	// 	srv := aws.NewService(ctx, serviceCode)
	// 	srv.GetPriceList()
	// }

	Test1(ctx)
}

func command() string {
	// Create flag
	srvFlag := flag.String("srv", "", "aws service code")
	flag.Parse()

	// Check flag
	if flag.NFlag() == 0 {
		return ""
	} else if *srvFlag != model.AWS_SERVICE_CODE_EC2 && *srvFlag != model.AWS_SERVICE_CODE_EBS && *srvFlag != model.AWS_SERVICE_CODE_RDS {
		return "Unknown"
	} else {
		return *srvFlag
	}
}

func Test1(ctx context.Context) {
	// Create service
	srv := aws.NewService(ctx, "AmazonVPC")
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
