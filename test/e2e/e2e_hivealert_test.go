/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

var _ = Describe("HiveAlert Simple", Ordered, func() {
	var err error

	var hiveTestPolicy = &hivev1alpha1.HivePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hive-policy-test1",
			Namespace: testNamespaceName,
			Finalizers: []string{hivev1alpha1.HivePolicyFinalizerName},
		},
		
		Spec: hivev1alpha1.HivePolicySpec{
			Path:   "/test",
			Create: true,
			Match: hivev1alpha1.HivePolicyMatch{
				PodName:   "test-pod",
				Namespace: "hive-test",
			},
		},
	}

	var testPod = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: testNamespaceName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test-pod",
				Image: "nginx:latest",
				Ports: []corev1.ContainerPort{{
					ContainerPort: 80,
				}},
			}},
		},
	}

	BeforeAll(func() {
		err = CleanHivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = CleanTestPods(ctx, Client, []corev1.Pod{testPod})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		err = CleanTestPods(ctx, Client, []corev1.Pod{testPod})
		Expect(err).NotTo(HaveOccurred())
		err = CleanHivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Operator", func() {

		It("Should not have any HivePolicy", func() {

			By("Getting HivePolicy")
			var hivePolicyList hivev1alpha1.HivePolicyList
			err := Client.List(ctx, &hivePolicyList, client.InNamespace(testNamespaceName))
			Expect(err).NotTo(HaveOccurred())

			if len(hivePolicyList.Items) != 0 {
				Expect(fmt.Errorf("HivePolicy present")).NotTo(HaveOccurred())
			}
		})
		It("Should not have any HiveData", func() {

			By("Getting HiveData")
			var hiveDataList hivev1alpha1.HiveDataList
			err := Client.List(ctx, &hiveDataList, client.InNamespace(operatorNamespace))
			Expect(err).NotTo(HaveOccurred())

			if len(hiveDataList.Items) != 0 {
				Expect(fmt.Errorf("HiveData present")).NotTo(HaveOccurred())
			}
		})
		It("Should succesfully create an HivePolicy", func() {

			By("Creating HivePolicy")
			err = Client.Create(ctx, hiveTestPolicy)
			Expect(err).NotTo(HaveOccurred())

			// Give the operator some time to react
			time.Sleep(reconcileTimeout)

			By("Getting HivePolicy")
			var hivePolicyList hivev1alpha1.HivePolicyList
			err := Client.List(ctx, &hivePolicyList, client.InNamespace(testNamespaceName))
			Expect(err).NotTo(HaveOccurred())

			if len(hivePolicyList.Items) != 1 {
				Expect(fmt.Errorf("HivePolicy not present")).NotTo(HaveOccurred())
			}

			By("Getting HiveData")
			var hiveDataList hivev1alpha1.HiveDataList
			err = Client.List(ctx, &hiveDataList, client.InNamespace(operatorNamespace))
			Expect(err).NotTo(HaveOccurred())

			if len(hiveDataList.Items) != 0 {
				Expect(fmt.Errorf("HiveData should not be present")).NotTo(HaveOccurred())
			}
		})
		It("Should create an HiveData when a new pod matches the policy", func() {

			By("Creating test pod")
			err = Client.Create(ctx, &testPod)
			if err != nil {
				Expect(fmt.Errorf("Creating Test Pod: %w", err)).NotTo(HaveOccurred())
			}

			By("Waiting for pod cration")
			key := client.ObjectKeyFromObject(&testPod)
			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				var p corev1.Pod
				if err := Client.Get(ctx, key, &p); err != nil {
					Expect(fmt.Errorf("Get Pod Pod: %w", err)).NotTo(HaveOccurred())
				}

				if p.Status.Phase == corev1.PodRunning {
					break
				}

				if p.Status.Phase == corev1.PodFailed || p.Status.Phase == corev1.PodSucceeded {
					Expect(fmt.Errorf("Pod Terminated: %s", p.Status.Phase)).NotTo(HaveOccurred())
				}

				time.Sleep(1 * time.Second)
			}

			// Give the operator some time to react
			time.Sleep(reconcileTimeout)

			By("Getting HiveData")
			var hiveDataList hivev1alpha1.HiveDataList
			if err := Client.List(ctx, &hiveDataList, client.InNamespace(operatorNamespace)); err != nil {
				Expect(fmt.Errorf("List HiveData: %w", err)).NotTo(HaveOccurred())
			}
			if len(hiveDataList.Items) != 1 {
				Expect(fmt.Errorf("One HiveData should be present, found %d", len(hiveDataList.Items))).NotTo(HaveOccurred())
			}
		})

		sinceTime := time.Now()

		It("Should have created the file in the matched pod", func() {
			cmd := exec.Command("kubectl", "exec", "-n", testNamespaceName, testPod.Name, "--", "cat", hiveTestPolicy.Spec.Path)
			fmt.Printf("Executing: %s", cmd.String())
			Expect(cmd.Run()).NotTo(HaveOccurred())
		})
		It("Should have generated an HiveAlert", func() {

			maxIt := 10
			it := 0
			for ; it < maxIt; it++ {
				cmd := exec.Command("kubectl", "logs", "-n", operatorNamespace, "-l", "control-plane=manager", "--tail", "1000", "--since-time", sinceTime.Format(time.RFC3339))
				fmt.Printf("Executing: %s\n", cmd.String())
				out, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				if strings.Contains(string(out), "HiveAlert") {
					break
				}

				time.Sleep(1 * time.Second)
			}
			if it == maxIt {
				Expect(fmt.Errorf("Should have received an alert")).NotTo(HaveOccurred())
			}
		})
		It("Should delete Hivedata after deletion of HivePolicy", func() {

			By("Deleting the HivePolicy")
			err = Client.Delete(ctx, hiveTestPolicy)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(reconcileTimeout)

			By("Getting the HiveData")
			var hiveDataList hivev1alpha1.HiveDataList
			err := Client.List(ctx, &hiveDataList, client.InNamespace(operatorNamespace))
			Expect(err).NotTo(HaveOccurred())

			if len(hiveDataList.Items) != 0 {
				Expect(fmt.Errorf("HiveData present")).NotTo(HaveOccurred())
			}
		})
	})
})
