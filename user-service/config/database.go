package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

func (cfg Config) ConnectionPostgres() (*Postgres, error) {
	dbConn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := gorm.Open(postgres.Open(dbConn), &gorm.Config{})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConnections)

	return &Postgres{
		DB: db,
	}, nil

}

func (p *Postgres) Close() {
	db, err := p.DB.DB()
	if err != nil {
		return
	}
	db.Close()
}
