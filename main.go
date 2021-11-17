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

	srv := aws.NewService(ctx, "AmazonEC2")
	srv.GetPriceList()
}

func Test1(ctx context.Context) {
	// Create service
	srv := aws.NewService(ctx, "AmazonEC2")
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
	values, err := srv.GetAttributeValues("instanceType")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("*--- Attribute Values (instance type) ---*")
		fmt.Println(values)
		fmt.Println()
	}
}
