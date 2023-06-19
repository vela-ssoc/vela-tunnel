package main

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
)

type elasticClient struct {
	cli   *elastic.Client
	inter time.Duration
	ctx   context.Context
}

func (ec *elasticClient) Monitor() {
	go ec.monitor()
}

func (ec *elasticClient) monitor() {
	ticker := time.NewTicker(ec.inter)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ticker.C:
			ec.health(5 * time.Second)
		}
	}
}

func (ec *elasticClient) health(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ec.ctx, timeout)
	defer cancel()

	res, err := ec.cli.CatHealth().Do(ctx)
	if err != nil {
		log.Printf("获取 es 健康错误：%s", err)
		return
	}
	for _, row := range res {
		log.Printf("获取 es 健康状态：%s-%s-%s", row.Cluster, row.Status, row.ActiveShardsPercent)
	}

	var doc1 any
	var doc2 any
	docs := []elastic.BulkableRequest{
		elastic.NewBulkCreateRequest().Doc(doc1),
		elastic.NewBulkCreateRequest().Doc(doc2),
	}
	ec.cli.Bulk().Index("xxx-xxx").Add(docs...).Do(ctx)
}
