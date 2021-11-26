package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	svc *s3.Client

	bucketName    string
	directoryName string
)

/*
 * AWS s3 configuration
 * @param 		ctx {context.Context} context
 * @response	{error} error object (contain nil)
 */
func Configure(ctx context.Context) error {
	// Configuration for AWS
	if cfg, err := config.LoadDefaultConfig(ctx); err != nil {
		return err
	} else {
		// Create service client for aws s3
		svc = s3.NewFromConfig(cfg)
		return nil
	}
}

/*
 * Set aws s3 bucket and directory to store output
 * @param			bName {string} bucket name
 * @param			dName {string} directory name in bucket
 * @response	{error} error object (contain nil)
 */
func SetPath(bName string, dName string) error {
	// Set s3 bucket
	if bucketName = bName; bucketName == "" {
		return errors.New("Setting the AWS S3 bucket name is mandatory")
	}
	// Set directory in s3
	directoryName = dName
	return nil
}

/*
 * Upload object to aws s3
 * @param			ctx {context.Context} context
 * @param			filename {string} output file name
 * @param			data {interface{}} output data
 * @response	{error} error object (contain nil)
 */
func UploadOutput(ctx context.Context, filename string, data interface{}) error {
	// Transform to byte
	transformed, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Create io reader
	reader := bytes.NewBuffer(transformed)

	// Set input parameter
	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(path.Join(directoryName, filename)),
		Body:          reader,
		ContentLength: int64(len(transformed)),
	}
	// Put object
	_, err = svc.PutObject(ctx, input, s3.WithAPIOptions(v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware))
	return err
}
