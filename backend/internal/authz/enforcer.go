package authz

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewEnforcer(modelPath, databaseURL string) (*casbin.Enforcer, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is required")
	}
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open authz database: %w", err)
	}

	gormadapter.TurnOffAutoMigrate(db)
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("create casbin adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load casbin policy: %w", err)
	}

	enforcer.EnableAutoSave(true)
	return enforcer, nil
}

func NewMemoryEnforcer(modelPath string) (*casbin.Enforcer, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is required")
	}

	model, err := casbinmodel.NewModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("load in-memory model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(model)
	if err != nil {
		return nil, fmt.Errorf("create in-memory enforcer: %w", err)
	}

	return enforcer, nil
}
