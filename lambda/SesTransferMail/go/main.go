package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	s3Client  *s3.Client
	sesClient *ses.Client
	s3Bucket  string
	forwardTo string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	sesClient = ses.NewFromConfig(cfg)
	s3Bucket = os.Getenv("S3_BUCKET")
	forwardTo = os.Getenv("FORWARD_TO")
}

func sendMail(message []byte) error {
	input := &ses.SendRawEmailInput{
		Source: aws.String(forwardTo),
		Destinations: []string{
			forwardTo,
		},
		RawMessage: &types.RawMessage{
			Data: message,
		},
	}

	_, err := sesClient.SendRawEmail(context.Background(), input)
	return err
}

func lambdaHandler(ctx context.Context, event map[string]interface{}) {
	log.Println("Event:", event)

	records := event["Records"].([]interface{})
	sesRecord := records[0].(map[string]interface{})
	mail := sesRecord["ses"].(map[string]interface{})["mail"].(map[string]interface{})
	messageId := mail["messageId"].(string)

	resp, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(messageId),
	})
	if err != nil {
		log.Fatalf("unable to get object from S3, %v", err)
	}
	defer resp.Body.Close()

	rawMessage, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unable to read object body, %v", err)
	}

	if err := sendMail(rawMessage); err != nil {
		log.Fatalf("unable to send email, %v", err)
	}
}

func main() {
	lambda.Start(lambdaHandler)
}
