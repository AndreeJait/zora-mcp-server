package main

import (
	"context"

	"github.com/AndreeJait/zora-mcp-server/adapter/outbound"
	"github.com/AndreeJait/zora-mcp-server/config"
	portOutbound "github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func provideInfrastructure(c *dig.Container) {
	c.Provide(newDB)
	c.Provide(newGormDB)
	c.Provide(newRedisConn)
	c.Provide(newRedisClient)
	c.Provide(newMinIOStorage)
}

func newDB(cfg *config.AppConfig, cc *CleanupCollector) (*outbound.DB, error) {
	db, cleanup, err := outbound.ConnectSQL(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	cc.Add(cleanup)
	return db, nil
}

func newGormDB(db *outbound.DB) *gorm.DB {
	return db.GormDB
}

func newRedisConn(cfg *config.AppConfig, cc *CleanupCollector) (*outbound.RedisConn, error) {
	conn, cleanup, err := outbound.ConnectRedis(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	cc.Add(cleanup)
	return conn, nil
}

func newRedisClient(conn *outbound.RedisConn) *redis.Client {
	return conn.Client
}

func newMinIOStorage(cfg *config.AppConfig) (portOutbound.Storage, error) {
	return outbound.NewMinIOStorage(cfg)
}
