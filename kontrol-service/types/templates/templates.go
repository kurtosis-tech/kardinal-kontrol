package templates

import (
	"fmt"
	"regexp"
	"strings"

	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

type Template struct {
	Template    []corev1.Service
	Description *string
	Name        string
	Id          string
}

func NewTemplate(services []corev1.Service, description *string, name string, id string) Template {
	return Template{
		Template:    services,
		Description: description,
		Name:        name,
		Id:          id,
	}
}

func (t *Template) ApplyTemplateOverrides(serviceConfigs []apitypes.ServiceConfig, templateSpec *apitypes.TemplateSpec) []apitypes.ServiceConfig {
	if templateSpec == nil {
		return serviceConfigs
	}

	args := templateSpec.Arguments

	logrus.Infof("Processing template '%v' with args '%v'", templateSpec.TemplateName, args)

	for i, serviceConfig := range serviceConfigs {
		for _, templateService := range t.Template {
			if templateService.Name != serviceConfig.Service.Name {
				continue
			}
			logrus.Infof("Found overrides for service '%s' in template '%s'", templateService.Name, t.Name)
			for key, value := range templateService.Annotations {
				if strings.HasPrefix(key, "kardinal.dev.service/") {
					if serviceConfig.Service.Annotations == nil {
						serviceConfig.Service.Annotations = make(map[string]string)
					}

					// Process the value for variable substitutions
					processedValue := processVariables(value, args)

					serviceConfig.Service.Annotations[key] = processedValue
				}
			}
		}
		serviceConfigs[i] = serviceConfig
	}

	return serviceConfigs
}

// TODO perhaps the argument passing should also work for default template
// ex. we have args in the main manifest; but no template but we can still provide args
func processVariables(value string, args *map[string]interface{}) string {
	re := regexp.MustCompile(`\$\{(\w+)(?::-([^}]*))?\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		parts := re.FindStringSubmatch(match)
		varName := parts[1]
		defaultValue := parts[2]

		if args != nil {
			if argValue, ok := (*args)[varName]; ok {
				switch v := argValue.(type) {
				case string:
					return v
				default:
					return fmt.Sprint(v)
				}
			}
		}
		return defaultValue
	})
}
