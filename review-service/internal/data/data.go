package data

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"review-service/internal/conf"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB)

// Data .
type Data struct {
	// TODO wrapped database client
	query *query.Query
}

// NewData .
func NewData(db *gorm.DB, c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	db, err := NewDB(c)
	if err != nil {
		panic(err)
	}
	query.SetDefault(db)
	return &Data{query: query.Q}, cleanup, nil
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(c.Database.GetSource()))
	if err != nil {
		panic(fmt.Errorf("connect db fail: %w", err))
	}
	return db, nil
}
