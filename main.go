package main

import (
	"aws-price-scanner/aws"
	"context"
	"fmt"
	"log"
)

func init() {
	// Configure an AWS SDK
	if err := aws.Configure(context.TODO()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.TODO()

	srv := aws.NewService(ctx, "AmazonRDS")
	srv.GetPriceList()

	// Test1(ctx)
}

func Test1(ctx context.Context) {
	// Create service
	srv := aws.NewService(ctx, "AmazonRDS")
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
