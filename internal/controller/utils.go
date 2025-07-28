package controller

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

func HiveDataPolicyCmp(hiveData hivev1alpha1.HiveData, hivePolicy hivev1alpha1.HivePolicy) bool {

	if (hiveData.Annotations["path"] != hivePolicy.Spec.Path) {
		return false
	}
	if hivePolicy.Spec.Match.PodName != "" &&
		hiveData.Annotations["pod_name"] != hivePolicy.Spec.Match.PodName {
		return false
	}
	if hivePolicy.Spec.Match.Namespace != "" &&
		hiveData.Annotations["namespace"] != hivePolicy.Spec.Match.Namespace {
		return false
	}
	if hivePolicy.Spec.Match.IP != "" &&
		hiveData.Annotations["pod_ip"] != hivePolicy.Spec.Match.IP {
		return false
	}

	sameLabels := true
	for label, value := range hivePolicy.Spec.Match.MatchLabels {
		hiveDataValue, ok := hiveData.Annotations["match-label-"+label]
		if !ok || value != hiveDataValue {
			sameLabels = false
			break
		}
	}

	return sameLabels
}

func HiveDataPodCmp(hiveData hivev1alpha1.HiveData, pod corev1.Pod) bool {

	if hiveData.Annotations["pod_name"] != pod.Name {
		return false
	}
	if hiveData.Annotations["namespace"] != pod.Namespace {
		return false
	}

	sameLabels := true
	for label, value := range hiveData.Annotations {
		if strings.HasPrefix(label, "match-label-") {
			podValue, ok := pod.Labels[strings.TrimLeft(label, "match-label-")]
			if !ok || value != podValue {
				sameLabels = false
				break
			}
		}
	}

	return sameLabels
}
