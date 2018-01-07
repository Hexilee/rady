package postgres

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
	"os"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	ry "rady"
)

type GormPostgresConfig struct {
	ry.Configuration
	App *ry.Application
}

type GormPostgresParameter struct {
	ry.Parameter
	Host     *string `value:"rady.postgres.host" default:"127.0.0.1"`
	Port     *string `value:"rady.postgres.port" default:"3306"`
	Database *string `value:"rady.postgres.database"`
	Username *string `value:"rady.postgres.username"`
	Password *string `value:"rady.postgres.password"`
	SSLMode  *string `value:"rady.postgres.sslmode" default:"disable"`
}

type GormPostgresRepo struct {
	ry.Repository
	Db *gorm.DB
}

func (g *GormPostgresConfig) GetAutoMigratePostgresDB(params *GormPostgresParameter) *GormPostgresRepo {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s", *params.Host, *params.Port, *params.Username, *params.SSLMode, *params.Database, *params.Password))
	if err != nil {
		g.App.Logger.Critical("Cannot connect to postgres \nError:\n%s", err.Error())
		os.Exit(1)
	}
	for _, entityType := range g.App.Entities {
		if entityType.Kind() == reflect.Ptr && entityType.Elem().Kind() == reflect.Struct {
			g.App.Logger.Debug("AutoMigrate: %s", entityType.String())
			db.AutoMigrate(reflect.New(entityType.Elem()).Interface())
		}
	}
	return &GormPostgresRepo{Db: db}
}