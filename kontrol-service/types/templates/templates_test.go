package templates

import (
	apitypes "github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api/api/golang/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestApplyTemplateOverrides(t *testing.T) {
	tests := []struct {
		name           string
		template       *Template
		serviceConfigs []apitypes.ServiceConfig
		templateSpec   *apitypes.TemplateSpec
		expected       []apitypes.ServiceConfig
	}{
		{
			name: "Nil TemplateSpec",
			template: &Template{
				Name: "test-template",
				Template: []corev1.Service{
					{ObjectMeta: metav1.ObjectMeta{Name: "service1"}},
				},
			},
			serviceConfigs: []apitypes.ServiceConfig{
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service1"}}},
			},
			templateSpec: nil,
			expected: []apitypes.ServiceConfig{
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service1"}}},
			},
		},
		{
			name: "Apply Annotations",
			template: &Template{
				Name: "test-template",
				Template: []corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation1": "value1",
								"kardinal.dev.service/annotation2": "${var:-default}",
							},
						},
					},
				},
			},
			serviceConfigs: []apitypes.ServiceConfig{
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service1"}}},
			},
			templateSpec: &apitypes.TemplateSpec{
				TemplateName: "test-template",
				Arguments:    &map[string]interface{}{"var": "custom"},
			},
			expected: []apitypes.ServiceConfig{
				{
					Service: corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation1": "value1",
								"kardinal.dev.service/annotation2": "custom",
							},
						},
					},
				},
			},
		},
		{
			name: "Multiple Services",
			template: &Template{
				Name: "test-template",
				Template: []corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation1": "${var1:-default1}",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service2",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation2": "${var2:-default2}",
							},
						},
					},
				},
			},
			serviceConfigs: []apitypes.ServiceConfig{
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service1"}}},
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service2"}}},
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service3"}}},
			},
			templateSpec: &apitypes.TemplateSpec{
				TemplateName: "test-template",
				Arguments:    &map[string]interface{}{"var1": "value1"},
			},
			expected: []apitypes.ServiceConfig{
				{
					Service: corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation1": "value1",
							},
						},
					},
				},
				{
					Service: corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service2",
							Annotations: map[string]string{
								"kardinal.dev.service/annotation2": "default2",
							},
						},
					},
				},
				{Service: corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "service3"}}},
			},
		},
		{
			name: "Replace Existing Annotation",
			template: &Template{
				Name: "test-template",
				Template: []corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/existing": "${newValue:-updatedDefault}",
							},
						},
					},
				},
			},
			serviceConfigs: []apitypes.ServiceConfig{
				{
					Service: corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/existing": "oldValue",
								"kardinal.dev.service/other":    "unchanged",
							},
						},
					},
				},
			},
			templateSpec: &apitypes.TemplateSpec{
				TemplateName: "test-template",
				Arguments:    &map[string]interface{}{"newValue": "replacedValue"},
			},
			expected: []apitypes.ServiceConfig{
				{
					Service: corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service1",
							Annotations: map[string]string{
								"kardinal.dev.service/existing": "replacedValue",
								"kardinal.dev.service/other":    "unchanged",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.template.ApplyTemplateOverrides(tt.serviceConfigs, tt.templateSpec)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ApplyTemplateOverrides() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessVariables(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		args     *map[string]interface{}
		expected string
	}{
		{
			name:     "Nil args",
			value:    "Hello ${name:-World}!",
			args:     nil,
			expected: "Hello World!",
		},
		{
			name:     "Empty args",
			value:    "Hello ${name:-World}!",
			args:     &map[string]interface{}{},
			expected: "Hello World!",
		},
		{
			name:     "Args with string value",
			value:    "Hello ${name:-World}!",
			args:     &map[string]interface{}{"name": "John"},
			expected: "Hello John!",
		},
		{
			name:     "Args with non-string value",
			value:    "The answer is ${answer:-unknown}.",
			args:     &map[string]interface{}{"answer": 42},
			expected: "The answer is 42.",
		},
		{
			name:     "Multiple substitutions",
			value:    "${greeting:-Hello}, ${name:-User}! Your score is ${score:-0}.",
			args:     &map[string]interface{}{"greeting": "Hi", "score": 100},
			expected: "Hi, User! Your score is 100.",
		},
		{
			name:     "No substitution needed",
			value:    "Plain text without variables",
			args:     &map[string]interface{}{"unused": "value"},
			expected: "Plain text without variables",
		},
		{
			name:     "Empty default value",
			value:    "Value: ${key:-}",
			args:     nil,
			expected: "Value: ",
		},
		{
			name:     "Variable without default",
			value:    "Name: ${name}",
			args:     nil,
			expected: "Name: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processVariables(tt.value, tt.args)
			if result != tt.expected {
				t.Errorf("processVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}
