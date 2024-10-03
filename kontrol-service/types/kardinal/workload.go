package types

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type Workload struct {
	Deployment  *appsv1.Deployment  `json:"deployment"`
	StatefulSet *appsv1.StatefulSet `json:"stateful_set"`
}

func (w *Workload) IsStatefulSet() bool {
	return w.StatefulSet != nil
}

func (w *Workload) IsDeployment() bool {
	return w.Deployment != nil
}

func (w *Workload) GetDeployment() *appsv1.Deployment {
	return w.Deployment
}

func (w *Workload) GetStatefulSet() *appsv1.StatefulSet {
	return w.StatefulSet
}

func NewDeploymentWorkload(deployment *appsv1.Deployment) Workload {
	return Workload{
		Deployment: deployment,
	}
}

func NewStatefulSetWorkload(statefulSet *appsv1.StatefulSet) Workload {
	return Workload{
		StatefulSet: statefulSet,
	}
}

func (w *Workload) DeepCopy() *Workload {
	if w == nil {
		return nil
	}

	if w.IsDeployment() {
		workload := NewDeploymentWorkload(w.GetDeployment().DeepCopy())
		return &workload
	} else if w.IsStatefulSet() {
		workload := NewStatefulSetWorkload(w.GetStatefulSet().DeepCopy())
		return &workload
	} else {
		panic("Invalid Workload")
	}
}

func (w *Workload) GetTemplateSpec() v1.PodSpec {
	if w.IsDeployment() {
		return w.GetDeployment().Spec.Template.Spec
	} else if w.IsStatefulSet() {
		return w.GetStatefulSet().Spec.Template.Spec
	} else {
		panic("Invalid Workload")
	}
}
