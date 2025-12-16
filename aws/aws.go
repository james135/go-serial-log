package aws

import (
	"bytes"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Session *s3.S3

func Configure(profile string) {
	os.Setenv("AWS_PROFILE", profile)
	config := &aws.Config{
		Region: aws.String("eu-west-1"),
	}
	session := session.Must(session.NewSession(config))
	S3Session = s3.New(session)
}

func ListCommonPrefixes(bucket, prefix string, delimeter string) ([]*s3.CommonPrefix, error) {
	var allContents []*s3.CommonPrefix
	var token *string
	for {
		listObjectsInput := s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
			Delimiter:         aws.String(delimeter),
		}
		objectList, err := S3Session.ListObjectsV2(&listObjectsInput)
		if err != nil {
			return nil, err
		}
		allContents = append(allContents, objectList.CommonPrefixes...)
		token = objectList.NextContinuationToken
		if token == nil {
			break
		}
	}
	return allContents, nil
}

func ListAllS3Objects(bucket, prefix string) ([]*s3.Object, error) {
	var allContents []*s3.Object
	var token *string
	for {
		listObjectsInput := s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		}
		objectList, err := S3Session.ListObjectsV2(&listObjectsInput)
		if err != nil {
			return nil, err
		}
		allContents = append(allContents, objectList.Contents...)
		token = objectList.NextContinuationToken
		if token == nil {
			break
		}
	}
	return allContents, nil
}

func DownloadFromS3(bucket string, object *s3.Object) ([]byte, error) {

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    object.Key,
	}
	result, err := S3Session.GetObject(input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(result.Body)
	return buffer.Bytes(), nil
}

func UploadToS3(bucket, key string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	}
	_, err := S3Session.PutObject(input)
	return err
}

func DeleteFromS3(bucket string, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err := S3Session.DeleteObject(input)
	return err
}
