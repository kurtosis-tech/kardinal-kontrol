package database

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Flow struct {
	gorm.Model
	FlowId          string `gorm:"uniqueIndex:idx_tenant_flow"`
	ClusterTopology datatypes.JSON
	TenantId        string `gorm:"uniqueIndex:idx_tenant_flow"`
}

func (db *Db) CreateFlow(
	tenantId string,
	flowId string,
	clusterTopology []byte,
) (*Flow, error) {
	flow := &Flow{
		FlowId:          flowId,
		ClusterTopology: datatypes.JSON(clusterTopology),
		TenantId:        tenantId,
	}
	result := db.db.Create(flow)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred creating the flow '%v'", flowId)
	}
	logrus.Infof("Success! Stored flow %s in database", flowId)
	return flow, nil
}

func (db *Db) DeleteFlow(
	tenantId string,
	flowId string,
) error {
	result := db.db.Where("tenant_id = ? AND flow_id = ?", tenantId, flowId).Delete(&Flow{})
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred deleting the flow '%v'", flowId)
	}
	logrus.Infof("Success! Deleted flow %s in database", flowId)
	return nil
}
