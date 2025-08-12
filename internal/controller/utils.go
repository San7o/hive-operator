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

	kivev2alpha1 "github.com/San7o/kivebpf/api/v2alpha1"
	container "github.com/San7o/kivebpf/internal/controller/container"
)

func KiveTrapHashID(kiveTrap kivev2alpha1.KiveTrap, alertVersion string) (string, error) {

	jsonTrap, err := json.Marshal(kiveTrap)
	if err != nil {
		return "", fmt.Errorf("TrapHashID Error Json Marshal: %w", err)
	}

	sha := sha256.New()
	sha.Write(jsonTrap)
	sha.Write([]byte(alertVersion))
	shaPolicy := hex.EncodeToString(sha.Sum(nil))

	return shaPolicy[:63], nil
}

func NewKiveDataName(inode uint64, dev uint32, pod corev1.Pod, containerStatus corev1.ContainerStatus) string {

	_, containerID, _ := container.SplitContainerRuntimeID(containerStatus.ContainerID)
	return strconv.FormatUint(inode, 10) +
		"-" +
		strconv.FormatUint(uint64(dev), 10) +
		"-kive-data-" +
		pod.Name +
		"-" +
		containerID
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

func KiveDataTrapCmp(kiveData kivev2alpha1.KiveData, kiveTrap kivev2alpha1.KiveTrap) (bool, error) {

	trapID, err := KiveTrapHashID(kiveTrap, kiveData.ObjectMeta.Annotations["kive-alert-version"])
	if err != nil {
		return false, fmt.Errorf("KiveDataTrapCmp Error Hash ID: %w", err)
	}
	return kiveData.ObjectMeta.Labels[TrapIdLabel] == trapID, nil
}

func KiveDataContainerCmp(kiveData kivev2alpha1.KiveData, pod corev1.Pod, containerStatus corev1.ContainerStatus) bool {

	if kiveData.Annotations["pod-name"] != pod.Name {
		return false
	}
	if kiveData.Annotations["namespace"] != pod.Namespace {
		return false
	}
	if kiveData.Annotations["container-name"] != containerStatus.Name {
		return false
	}
	if kiveData.Annotations["container-id"] != containerStatus.ContainerID {
		return false
	}

	return true
}
