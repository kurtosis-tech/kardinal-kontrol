package types

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type WorkloadSpec struct {
	DeploymentSpec  *appsv1.DeploymentSpec  `json:"deployment_spec"`
	StatefulSetSpec *appsv1.StatefulSetSpec `json:"stateful_set_spec"`
}

func (w *WorkloadSpec) IsStatefulSet() bool {
	return w != nil && w.StatefulSetSpec != nil
}

func (w *WorkloadSpec) IsDeployment() bool {
	return w != nil && w.DeploymentSpec != nil
}

func (w *WorkloadSpec) GetDeploymentSpec() *appsv1.DeploymentSpec {
	return w.DeploymentSpec
}

func (w *WorkloadSpec) GetStatefulSetSpec() *appsv1.StatefulSetSpec {
	return w.StatefulSetSpec
}

func NewDeploymentWorkloadSpec(spec appsv1.DeploymentSpec) WorkloadSpec {
	return WorkloadSpec{
		DeploymentSpec: &spec,
	}
}

func NewStatefulSetWorkloadSpec(spec appsv1.StatefulSetSpec) WorkloadSpec {
	return WorkloadSpec{
		StatefulSetSpec: &spec,
	}
}

func (w *Workload) WorkloadSpec() *WorkloadSpec {
	if w == nil {
		return nil
	}

	if w.IsDeployment() {
		spec := NewDeploymentWorkloadSpec(w.GetDeployment().Spec)
		return &spec
	} else if w.IsStatefulSet() {
		spec := NewStatefulSetWorkloadSpec(w.GetStatefulSet().Spec)
		return &spec
	} else {
		panic("Invalid workload")
	}
}

func (w *WorkloadSpec) DeepCopy() *WorkloadSpec {
	if w == nil {
		return nil
	}

	if w.IsDeployment() {
		spec := NewDeploymentWorkloadSpec(*w.GetDeploymentSpec().DeepCopy())
		return &spec
	} else if w.IsStatefulSet() {
		spec := NewStatefulSetWorkloadSpec(*w.GetStatefulSetSpec().DeepCopy())
		return &spec
	} else {
		panic("Invalid WorkloadSpec")
	}
}

func (w *WorkloadSpec) GetTemplateSpec() *v1.PodSpec {
	if w.IsDeployment() {
		return &w.GetDeploymentSpec().Template.Spec
	} else if w.IsStatefulSet() {
		return &w.GetStatefulSetSpec().Template.Spec
	}

	return nil
}

func (w *WorkloadSpec) UpdateTemplateSpec(spec v1.PodSpec) {
	if w.IsDeployment() {
		w.GetDeploymentSpec().Template.Spec = spec
	} else if w.IsStatefulSet() {
		w.GetStatefulSetSpec().Template.Spec = spec
	}
}
