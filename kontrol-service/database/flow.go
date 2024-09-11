package database

import (
	"errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	// TODO retro compatibility value, we should remove this when we are sure that none is using it
	// TODO this was the first name of the default baseline which was used for all the baseline flows at the beginning, now we use the namespace value
	oldDefaultBaselineFlowId = "prod"
)

type Flow struct {
	gorm.Model
	FlowId          string `gorm:"uniqueIndex:idx_tenant_flow"`
	ClusterTopology datatypes.JSON
	TenantId        string `gorm:"uniqueIndex:idx_tenant_flow"`
	IsBaseline      bool
}

func (f *Flow) IsBaselineFlow() bool {
	// TODO retro compatibility value, we should remove this when we are sure that none is using it
	if f.FlowId == oldDefaultBaselineFlowId {
		return true
	}
	return f.IsBaseline
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

func (db *Db) GetFlow(flowId string) (*Flow, error) {
	var flow Flow
	result := db.db.Where("flow_id = ?", flowId).First(&flow)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred getting the flow '%v'", flowId)
	}
	return &flow, nil
}

func (db *Db) GetFlows() ([]Flow, error) {
	var flows []Flow
	result := db.db.Find(&flows)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred getting all flows")
	}
	return flows, nil
}

func (db *Db) GetBaselineFlow() (*Flow, error) {
	var flow Flow
	result := db.db.Where("is_baseline = ?", true).First(&flow)
	if result.Error != nil {
		// TODO retro compatibility implementation, we should remove this when we are sure that none is using it
		// TODO we infer the baselineFlow if flowID = "prod" which was the first implementation for baseline flow
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			allFlows, err := db.GetFlows()
			if err != nil {
				return nil, stacktrace.Propagate(err, "An internal error has occurred getting all flows to infer the baseline flow")
			}
			for _, flowObj := range allFlows {
				if flowObj.IsBaselineFlow() {
					return &flowObj, nil
				}
			}
		}
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred getting the baseline flow")
	}
	return &flow, nil
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

func (db *Db) DeleteTenantFlows(
	tenantId string,
) error {
	result := db.db.Where("tenant_id = ?", tenantId).Delete(&Flow{})
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred deleting the tenants %s flows", tenantId)
	}
	logrus.Infof("Success! Deleted tenant %s flows in database", tenantId)
	return nil
}
