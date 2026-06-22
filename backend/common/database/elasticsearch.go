package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"

	"github.com/olivere/elastic/v7"
)

var (
	esClient *elastic.Client
	esOnce   sync.Once
)

func InitES(cfg *config.ESConfig) *elastic.Client {
	esOnce.Do(func() {
		if len(cfg.Addresses) == 0 {
			logger.Error("Elasticsearch addresses not configured")
			return
		}

		opts := []elastic.ClientOptionFunc{
			elastic.SetURL(cfg.Addresses...),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(true),
		}

		if cfg.Username != "" {
			opts = append(opts, elastic.SetBasicAuth(cfg.Username, cfg.Password))
		}

		var err error
		esClient, err = elastic.NewClient(opts...)
		if err != nil {
			logger.Error("Initialize Elasticsearch client failed", "error", err)
			return
		}

		info, code, err := esClient.Ping(cfg.Addresses[0]).Do(context.Background())
		if err != nil {
			logger.Error("Ping Elasticsearch failed", "error", err)
			return
		}

		logger.Info("Elasticsearch client initialized",
			"status", code,
			"version", info.Version.Number,
		)
	})
	return esClient
}

func GetESClient() *elastic.Client {
	return esClient
}

func EnsureESIndex(indexName string, mapping string) error {
	if esClient == nil {
		return fmt.Errorf("elasticsearch client not initialized")
	}

	ctx := context.Background()
	exists, err := esClient.IndexExists(indexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("check index exists failed: %w", err)
	}

	if !exists {
		_, err = esClient.CreateIndex(indexName).Body(mapping).Do(ctx)
		if err != nil {
			return fmt.Errorf("create index failed: %w", err)
		}
		logger.Info("Elasticsearch index created", "index", indexName)
	}

	return nil
}
