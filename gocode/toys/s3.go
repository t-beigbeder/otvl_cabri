package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

func main() {
	epr := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL: "https://s3.gra.cloud.ovh.net",
		}, nil
	}
	config := aws.Config{
		Credentials:      credentials.NewStaticCredentials(os.Getenv("OVHAK"), os.Getenv("OVHSK"), ""),
		Region:           aws.String("GRA"),
		EndpointResolver: endpoints.ResolverFunc(epr),
	}
	sess := session.Must(session.NewSession(&config))

	s3Svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String("ooca"),
		Key:    aws.String("775px-Debian-OpenLogo.svg.png"),
	}
	result, err := s3Svc.GetObject(input)
	fmt.Fprintf(os.Stderr, "%+v %+v\n", result, err)

	// upload
	uploader := s3manager.NewUploader(sess)
	filename := "/home/guest/Documents/s3.go"
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s %v\n", filename, err)
		os.Exit(1)
	}
	result2, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("ooca"),
		Key:    aws.String("s3.go"),
		Body:   f,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "upload %s %v\n", filename, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%+v %+v\n", result2, err)

	fo, err := os.Create("/tmp/tbe.s3.go")
	downloader := s3manager.NewDownloader(sess)
	n, err := downloader.Download(fo, &s3.GetObjectInput{
		Bucket: aws.String("ooca"),
		Key:    aws.String("s3.go"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "download %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%+v %+v\n", n, err)

	s3Svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String("ooca"),
		Prefix: aws.String("sub"),
	}, func(output *s3.ListObjectsV2Output, b bool) bool {
		fmt.Fprintf(os.Stderr, "%+v %v\n", output, b)
		return true
	})

	s3Svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String("ooca"),
	}, func(output *s3.ListObjectsV2Output, b bool) bool {
		fmt.Fprintf(os.Stderr, "%+v %v\n", output, b)
		return true
	})

	s3Svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String("ooca"),
		Prefix: aws.String("none"),
	}, func(output *s3.ListObjectsV2Output, b bool) bool {
		fmt.Fprintf(os.Stderr, "%+v %v\n", output, b)
		return true
	})

	os.Exit(0)
}
