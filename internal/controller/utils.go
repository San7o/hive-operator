/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	container "github.com/San7o/hive-operator/internal/controller/container"
)

func HiveTrapHashID(hiveTrap hivev1alpha1.HiveTrap) (string, error) {

	jsonPolicy, err := json.Marshal(hiveTrap)
	if err != nil {
		return "", fmt.Errorf("TrapHashID Error Json Marshal: %w", err)
	}

	sha := sha256.New()
	sha.Write(jsonPolicy)
	shaPolicy := hex.EncodeToString(sha.Sum(nil))

	return shaPolicy[:63], nil
}

func NewHiveDataName(inode uint64, containerStatus corev1.ContainerStatus) string {

	_, containerID, _ := container.SplitContainerRuntimeID(containerStatus.ContainerID)
	return strconv.FormatUint(inode, 10) + "-hive-data-" + containerID
}

func RegexMatch(regex string, containerName string) (bool, error) {

	if regex == "" {
		return true, nil
	}

	compiledRegex, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("RegexMatch Error compiling regex: %w", err)
	}

	return compiledRegex.Match([]byte(containerName)), nil
}

func HiveDataTrapCmp(hiveData hivev1alpha1.HiveData, hiveTrap hivev1alpha1.HiveTrap) (bool, error) {

	trapID, err := HiveTrapHashID(hiveTrap)
	if err != nil {
		return false, fmt.Errorf("HiveDataTrapCmp Error Hash ID: %w", err)
	}
	return hiveData.ObjectMeta.Labels[TrapIdLabel] == trapID, nil
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
	if hiveData.Annotations["container_id"] != containerStatus.ContainerID {
		return false
	}

	return true
}
