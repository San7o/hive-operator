package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	container "github.com/San7o/hive-operator/internal/controller/container"
)

func PolicyHashID(hivePolicy hivev1alpha1.HivePolicy) (string, error) {

	jsonPolicy, err := json.Marshal(hivePolicy.Spec.Match)
	if err != nil {
		return "", fmt.Errorf("PolicyHashID Error Json Marshal: %w", err)
	}

	sha := sha256.New()
	sha.Write(append(jsonPolicy, []byte(hivePolicy.Spec.Path)...))
	shaPolicy := hex.EncodeToString(sha.Sum(nil))

	return shaPolicy[:63], nil
}

func NewHiveDataName(inode uint64, containerStatus corev1.ContainerStatus) string {
	
	_, containerID, _ := container.SplitContainerRuntimeID(containerStatus.ContainerID)	
	return strconv.FormatUint(inode, 10) + "-hive-data-" + containerID
}

func HiveDataPolicyCmp(hiveData hivev1alpha1.HiveData, hivePolicy hivev1alpha1.HivePolicy) bool {

	return hiveData.ObjectMeta.Labels["policy-id"] == hivePolicy.ObjectMeta.Labels["policy-id"]
}

func HiveDataContainerCmp(hiveData hivev1alpha1.HiveData, pod corev1.Pod, containerStatus corev1.ContainerStatus) bool {

	if hiveData.Annotations["pod_name"] != pod.Name {
		return false
	}
	if hiveData.Annotations["namespace"] != pod.Namespace {
		return false
	}
	if hiveData.Annotations["container_name"] != containerStatus.Name {
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
