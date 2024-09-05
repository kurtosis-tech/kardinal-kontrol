package database

import (
	"github.com/kurtosis-tech/stacktrace"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	TenantId            string `gorm:"uniqueIndex"`
	BaseClusterTopology datatypes.JSON
	ServiceConfigs      datatypes.JSON
	IngressConfigs      datatypes.JSON
	Active              bool
	Flows               []Flow         `gorm:"foreignKey:TenantId;references:TenantId;constraint:OnDelete:CASCADE"`
	PluginConfigs       []PluginConfig `gorm:"foreignKey:TenantId;references:TenantId;constraint:OnDelete:CASCADE"`
	Templates           []Template     `gorm:"foreignKey:TenantId;references:TenantId;constraint:OnDelete:CASCADE"`
}

func (db *Db) SaveTenant(tenant *Tenant) error {
	result := db.db.Save(tenant)
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred updating the tenant '%v'", tenant.TenantId)
	}
	return nil
}

func (db *Db) GetOrCreateTenant(
	tenantId string,
) (*Tenant, error) {
	var tenant Tenant
	active := true
	result := db.db.Where(Tenant{
		TenantId: tenantId,
		Active:   active,
	}).Preload("Flows").Preload("Templates").FirstOrCreate(&tenant)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "an error occurred while fetching tenant %s", tenantId)
	}

	return &tenant, nil
}

func (db *Db) GetTenant(
	tenantId string,
) (*Tenant, error) {
	var tenant Tenant
	result := db.db.Where("tenant_id = ?", tenantId).Preload("Flows").Preload("Templates").First(&tenant)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tenant, nil
}
