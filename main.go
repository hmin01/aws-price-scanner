package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	// Custom aws module
	"aws-price-scanner/aws/pricing"
	"aws-price-scanner/aws/s3"

	// Model
	"aws-price-scanner/model"
)

const (
	ENV_ServiceKey   = "serviceCode"
	ENV_BucketKey    = "bucket"
	ENV_DirectoryKey = "directory"
)

var testServiceCode = "AmazonECS"

func init() {
	// Configure an AWS pricing
	if err := pricing.Configure(context.TODO()); err != nil {
		log.Fatal(err)
	}
	// Configure an AWS S3
	if err := s3.Configure(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.TODO()

	// Create flag to use command line
	if err := command(); err != nil {
		fmt.Println(err.Error() + "\r\n")
		flag.Usage()
		os.Exit(100)
	} else {
		if serviceCode := os.Getenv(ENV_ServiceKey); serviceCode == "" {
			// Get a list of service code
			list, err := pricing.GetServiceCodeList(ctx)
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("*-- Service code list ---*")
				for _, elem := range list {
					fmt.Println(elem)
				}
			}
		} else {
			// Set s3
			if err := s3.SetPath(os.Getenv(ENV_BucketKey), os.Getenv(ENV_DirectoryKey)); err != nil {
				fmt.Println(err.Error())
				os.Exit(101)
			}
			// Process
			srv := pricing.NewService(ctx, serviceCode)
			srv.GetPriceList()
		}
	}

	// Test1(ctx)

	// Test2(ctx)
}

func command() error {
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
	bucketFlag := flag.String("bucket", "", "AWS S3 bucket name to store output")
	directoryFlag := flag.String("directory", "", "Directory path in AWS S3 bucket")
	flag.Parse()

	// Check flag
	if flag.NFlag() == 0 {
		os.Setenv(ENV_ServiceKey, "")
	} else {
		if *srvFlag != "" {
			match := false
			for _, code := range model.AWS_SERVICE_CODE_LIST {
				if code == *srvFlag {
					match = true
					break
				}
			}
			// Return
			if match {
				os.Setenv(ENV_ServiceKey, *srvFlag)
			} else {
				return errors.New("Not match service code")
			}
		}

		if *directoryFlag != "" {
			os.Setenv(ENV_DirectoryKey, *directoryFlag)
		}

		if *bucketFlag != "" {
			os.Setenv(ENV_BucketKey, *bucketFlag)
		} else {
			return errors.New("Storage paths for storing results are essential.")
		}
	}
	return nil
}

func Test1(ctx context.Context) {
	// Create service
	srv := pricing.NewService(ctx, testServiceCode)
	// Get attributes for service
	attributes, err := srv.GetAttributes()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("*--- Attributes ---*")
		length := len(attributes)
		for index, attribute := range attributes {
			if index < length-1 {
				fmt.Print(attribute + ", ")
			} else {
				fmt.Println(attribute)
			}
		}
		fmt.Println()
	}
	// Get attribute values for service
	for _, attribute := range attributes {
		values, err := srv.GetAttributeValues(attribute)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("*--- " + attribute + " ---*")
			length := len(values)
			for index, value := range values {
				if index < length-1 {
					fmt.Print(value + ", ")
				} else {
					fmt.Println(value)
				}
			}
			fmt.Println()
		}
	}
}

func Test2(ctx context.Context) {
	// Create service
	srv := pricing.NewService(ctx, testServiceCode)
	// Get price list (for test)
	if err := srv.GetPriceListForTest(); err != nil {
		log.Fatal(err)
	}
}
