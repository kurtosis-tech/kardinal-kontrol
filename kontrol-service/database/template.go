package database

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Template struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex:idx_tenant_template"`
	Body     datatypes.JSON
	TenantId string `gorm:"uniqueIndex:idx_tenant_template"`
}

func (db *Db) CreateTemplate(
	tenantId string,
	name string,
	template []byte,
) (*Template, error) {
	templateToCreate := &Template{
		Name:     name,
		Body:     datatypes.JSON(template),
		TenantId: tenantId,
	}
	result := db.db.Create(templateToCreate)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred creating the template '%v'", name)
	}
	logrus.Infof("Success! Stored template %s in database", name)
	return templateToCreate, nil
}

func (db *Db) SaveTemplate(template *Template) error {
	result := db.db.Save(template)
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred updating the template '%v'", template.Name)
	}
	return nil
}

func (db *Db) DeleteTemplate(
	tenantId string,
	name string,
) error {
	result := db.db.Where("tenant_id = ? AND name = ?", tenantId, name).Delete(&Template{})
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred deleting the template '%v'", name)
	}
	logrus.Infof("Success! Deleted template %s in database", name)
	return nil
}
