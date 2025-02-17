//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectMetadata) DeepCopyInto(out *ObjectMetadata) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectMetadata.
func (in *ObjectMetadata) DeepCopy() *ObjectMetadata {
	if in == nil {
		return nil
	}
	out := new(ObjectMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PVC) DeepCopyInto(out *PVC) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PVC.
func (in *PVC) DeepCopy() *PVC {
	if in == nil {
		return nil
	}
	out := new(PVC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpdateStrategyDelayedRollingUpdate) DeepCopyInto(out *UpdateStrategyDelayedRollingUpdate) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpdateStrategyDelayedRollingUpdate.
func (in *UpdateStrategyDelayedRollingUpdate) DeepCopy() *UpdateStrategyDelayedRollingUpdate {
	if in == nil {
		return nil
	}
	out := new(UpdateStrategyDelayedRollingUpdate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VCLStatus) DeepCopyInto(out *VCLStatus) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VCLStatus.
func (in *VCLStatus) DeepCopy() *VCLStatus {
	if in == nil {
		return nil
	}
	out := new(VCLStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishCluster) DeepCopyInto(out *VarnishCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishCluster.
func (in *VarnishCluster) DeepCopy() *VarnishCluster {
	if in == nil {
		return nil
	}
	out := new(VarnishCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VarnishCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterBackend) DeepCopyInto(out *VarnishClusterBackend) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ZoneBalancing != nil {
		in, out := &in.ZoneBalancing, &out.ZoneBalancing
		*out = new(VarnishClusterBackendZoneBalancing)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterBackend.
func (in *VarnishClusterBackend) DeepCopy() *VarnishClusterBackend {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterBackendZoneBalancing) DeepCopyInto(out *VarnishClusterBackendZoneBalancing) {
	*out = *in
	if in.Thresholds != nil {
		in, out := &in.Thresholds, &out.Thresholds
		*out = make([]VarnishClusterBackendZoneBalancingThreshold, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterBackendZoneBalancing.
func (in *VarnishClusterBackendZoneBalancing) DeepCopy() *VarnishClusterBackendZoneBalancing {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterBackendZoneBalancing)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterBackendZoneBalancingThreshold) DeepCopyInto(out *VarnishClusterBackendZoneBalancingThreshold) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = new(int)
		**out = **in
	}
	if in.Remote != nil {
		in, out := &in.Remote, &out.Remote
		*out = new(int)
		**out = **in
	}
	if in.Threshold != nil {
		in, out := &in.Threshold, &out.Threshold
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterBackendZoneBalancingThreshold.
func (in *VarnishClusterBackendZoneBalancingThreshold) DeepCopy() *VarnishClusterBackendZoneBalancingThreshold {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterBackendZoneBalancingThreshold)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterList) DeepCopyInto(out *VarnishClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VarnishCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterList.
func (in *VarnishClusterList) DeepCopy() *VarnishClusterList {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VarnishClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterMonitoring) DeepCopyInto(out *VarnishClusterMonitoring) {
	*out = *in
	if in.PrometheusServiceMonitor != nil {
		in, out := &in.PrometheusServiceMonitor, &out.PrometheusServiceMonitor
		*out = new(VarnishClusterMonitoringPrometheusServiceMonitor)
		(*in).DeepCopyInto(*out)
	}
	if in.GrafanaDashboard != nil {
		in, out := &in.GrafanaDashboard, &out.GrafanaDashboard
		*out = new(VarnishClusterMonitoringGrafanaDashboard)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterMonitoring.
func (in *VarnishClusterMonitoring) DeepCopy() *VarnishClusterMonitoring {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterMonitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterMonitoringGrafanaDashboard) DeepCopyInto(out *VarnishClusterMonitoringGrafanaDashboard) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DatasourceName != nil {
		in, out := &in.DatasourceName, &out.DatasourceName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterMonitoringGrafanaDashboard.
func (in *VarnishClusterMonitoringGrafanaDashboard) DeepCopy() *VarnishClusterMonitoringGrafanaDashboard {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterMonitoringGrafanaDashboard)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterMonitoringPrometheusServiceMonitor) DeepCopyInto(out *VarnishClusterMonitoringPrometheusServiceMonitor) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterMonitoringPrometheusServiceMonitor.
func (in *VarnishClusterMonitoringPrometheusServiceMonitor) DeepCopy() *VarnishClusterMonitoringPrometheusServiceMonitor {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterMonitoringPrometheusServiceMonitor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterService) DeepCopyInto(out *VarnishClusterService) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	if in.MetricsPort != nil {
		in, out := &in.MetricsPort, &out.MetricsPort
		*out = new(int32)
		**out = **in
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterService.
func (in *VarnishClusterService) DeepCopy() *VarnishClusterService {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterSpec) DeepCopyInto(out *VarnishClusterSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.UpdateStrategy != nil {
		in, out := &in.UpdateStrategy, &out.UpdateStrategy
		*out = new(VarnishClusterUpdateStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.Varnish != nil {
		in, out := &in.Varnish, &out.Varnish
		*out = new(VarnishClusterVarnish)
		(*in).DeepCopyInto(*out)
	}
	if in.VCL != nil {
		in, out := &in.VCL, &out.VCL
		*out = new(VarnishClusterVCL)
		(*in).DeepCopyInto(*out)
	}
	if in.Backend != nil {
		in, out := &in.Backend, &out.Backend
		*out = new(VarnishClusterBackend)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(VarnishClusterService)
		(*in).DeepCopyInto(*out)
	}
	if in.PodDisruptionBudget != nil {
		in, out := &in.PodDisruptionBudget, &out.PodDisruptionBudget
		*out = new(v1.PodDisruptionBudgetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodAnnotations != nil {
		in, out := &in.PodAnnotations, &out.PodAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Monitoring != nil {
		in, out := &in.Monitoring, &out.Monitoring
		*out = new(VarnishClusterMonitoring)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterSpec.
func (in *VarnishClusterSpec) DeepCopy() *VarnishClusterSpec {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterStatus) DeepCopyInto(out *VarnishClusterStatus) {
	*out = *in
	in.VCL.DeepCopyInto(&out.VCL)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterStatus.
func (in *VarnishClusterStatus) DeepCopy() *VarnishClusterStatus {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterUpdateStrategy) DeepCopyInto(out *VarnishClusterUpdateStrategy) {
	*out = *in
	if in.RollingUpdate != nil {
		in, out := &in.RollingUpdate, &out.RollingUpdate
		*out = new(appsv1.RollingUpdateStatefulSetStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.DelayedRollingUpdate != nil {
		in, out := &in.DelayedRollingUpdate, &out.DelayedRollingUpdate
		*out = new(UpdateStrategyDelayedRollingUpdate)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterUpdateStrategy.
func (in *VarnishClusterUpdateStrategy) DeepCopy() *VarnishClusterUpdateStrategy {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterUpdateStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterVCL) DeepCopyInto(out *VarnishClusterVCL) {
	*out = *in
	if in.ConfigMapName != nil {
		in, out := &in.ConfigMapName, &out.ConfigMapName
		*out = new(string)
		**out = **in
	}
	if in.EntrypointFileName != nil {
		in, out := &in.EntrypointFileName, &out.EntrypointFileName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterVCL.
func (in *VarnishClusterVCL) DeepCopy() *VarnishClusterVCL {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterVCL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterVarnish) DeepCopyInto(out *VarnishClusterVarnish) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Controller != nil {
		in, out := &in.Controller, &out.Controller
		*out = new(VarnishClusterVarnishController)
		(*in).DeepCopyInto(*out)
	}
	if in.MetricsExporter != nil {
		in, out := &in.MetricsExporter, &out.MetricsExporter
		*out = new(VarnishClusterVarnishMetricsExporter)
		(*in).DeepCopyInto(*out)
	}
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(VarnishClusterVarnishSecret)
		(*in).DeepCopyInto(*out)
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]corev1.EnvFromSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExtraInitContainers != nil {
		in, out := &in.ExtraInitContainers, &out.ExtraInitContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExtraVolumeClaimTemplates != nil {
		in, out := &in.ExtraVolumeClaimTemplates, &out.ExtraVolumeClaimTemplates
		*out = make([]PVC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExtraVolumes != nil {
		in, out := &in.ExtraVolumes, &out.ExtraVolumes
		*out = make([]corev1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExtraVolumeMounts != nil {
		in, out := &in.ExtraVolumeMounts, &out.ExtraVolumeMounts
		*out = make([]corev1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterVarnish.
func (in *VarnishClusterVarnish) DeepCopy() *VarnishClusterVarnish {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterVarnish)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterVarnishController) DeepCopyInto(out *VarnishClusterVarnishController) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterVarnishController.
func (in *VarnishClusterVarnishController) DeepCopy() *VarnishClusterVarnishController {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterVarnishController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterVarnishMetricsExporter) DeepCopyInto(out *VarnishClusterVarnishMetricsExporter) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterVarnishMetricsExporter.
func (in *VarnishClusterVarnishMetricsExporter) DeepCopy() *VarnishClusterVarnishMetricsExporter {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterVarnishMetricsExporter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishClusterVarnishSecret) DeepCopyInto(out *VarnishClusterVarnishSecret) {
	*out = *in
	if in.SecretName != nil {
		in, out := &in.SecretName, &out.SecretName
		*out = new(string)
		**out = **in
	}
	if in.Key != nil {
		in, out := &in.Key, &out.Key
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishClusterVarnishSecret.
func (in *VarnishClusterVarnishSecret) DeepCopy() *VarnishClusterVarnishSecret {
	if in == nil {
		return nil
	}
	out := new(VarnishClusterVarnishSecret)
	in.DeepCopyInto(out)
	return out
}
