package database

import (
	"context"
	"sync"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var (
	minioClient *minio.Client
	minioOnce   sync.Once
)

func InitMinIO(cfg *config.MinIOConfig) *minio.Client {
	minioOnce.Do(func() {
		var err error
		minioClient, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.UseSSL,
		})
		if err != nil {
			logger.Error("Initialize MinIO client failed", logger.Error(err))
			return
		}

		if cfg.Bucket != "" {
			exists, err := minioClient.BucketExists(context.Background(), cfg.Bucket)
			if err != nil {
				logger.Error("Check bucket exists failed",
					zap.String("bucket", cfg.Bucket),
					logger.Error(err),
				)
				return
			}
			if !exists {
				err = minioClient.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{})
				if err != nil {
					logger.Error("Create bucket failed",
						zap.String("bucket", cfg.Bucket),
						logger.Error(err),
					)
					return
				}
				logger.Info("Bucket created", zap.String("bucket", cfg.Bucket))
			}
		}

		logger.Info("MinIO client initialized",
			zap.String("endpoint", cfg.Endpoint),
			zap.String("bucket", cfg.Bucket),
		)
	})
	return minioClient
}

func GetMinioClient() *minio.Client {
	return minioClient
}
