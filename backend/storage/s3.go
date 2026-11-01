package storage

import (
	"github.com/minio/minio-go/v7/pkg/credentials"
	"Test-K8s-Pet-Project-Backend/configs"
	"github.com/minio/minio-go/v7"
	"context"
)

var Client *minio.Client

func InitS3(ctx context.Context) error {
	c, err := minio.New(configs.S3Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(configs.S3AccessKey(), configs.S3SecretKey(), ""),
		Secure: configs.S3UseSSL(),
		Region: configs.S3Region(),
	})
	if err != nil {
		return err
	}
	exists, err := c.BucketExists(ctx, configs.S3Bucket())
	if err != nil {
		return err
	}
	if !exists {
		if err := c.MakeBucket(ctx, configs.S3Bucket(), minio.MakeBucketOptions{Region: configs.S3Region()}); err != nil {
			return err
		}
	}
	Client = c
	return nil
}