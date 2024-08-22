package database

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PluginConfig struct {
	gorm.Model
	FlowId   string `gorm:"uniqueIndex:idx_tenant_plugin"`
	Config   string
	TenantId string `gorm:"uniqueIndex:idx_tenant_plugin"`
}

func (db *Db) CreatePluginConfig(
	flowId string,
	config string,
	tenantId string,
) (*PluginConfig, error) {
	pluginConfig := &PluginConfig{
		FlowId:   flowId,
		Config:   config,
		TenantId: tenantId,
	}
	result := db.db.Create(pluginConfig)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred creating the plugin config for flow '%v'", flowId)
	}
	logrus.Infof("Success! Stored plugin config for flow %s in database", flowId)
	return pluginConfig, nil
}

func (db *Db) DeletePluginConfig(
	tenantId string,
	flowId string,
) error {
	result := db.db.Where("tenant_id = ? AND flow_id = ?", tenantId, flowId).Delete(&PluginConfig{})
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred deleting the plugin config for flow '%v'", flowId)
	}

	return nil
}

func (db *Db) GetPluginConfigByFlowID(
	tenantId string,
	flowId string,
) (*PluginConfig, error) {
	var pluginConfig PluginConfig
	result := db.db.Where("tenant_id = ? AND flow_id = ?", tenantId, flowId).First(&pluginConfig)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &pluginConfig, nil
}
