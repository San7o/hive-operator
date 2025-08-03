/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

var _ = Describe("ContainerName Regex 2", Ordered, func() {
	var err error

	var hiveTestPolicy = &hivev1alpha1.HivePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "hive-policy-test-regex2",
			Namespace:  testNamespaceName,
			Finalizers: []string{hivev1alpha1.HivePolicyFinalizerName},
		},
		Spec: hivev1alpha1.HivePolicySpec{
			Traps: []hivev1alpha1.HiveTrap{
				{
					Path:   "/regex2",
					Create: true,
					MatchAny: []hivev1alpha1.HiveTrapMatch{
						hivev1alpha1.HiveTrapMatch{
							PodName:       "test-pod",
							Namespace:     "hive-test",
							ContainerName: "test-nope.*",
						},
					},
				},
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
				Name:  "test-nginx",
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

		It("Should not create an HiveData when a pod does not match the container name", func() {

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
			if len(hiveDataList.Items) != 0 {
				Expect(fmt.Errorf("One HiveData should not be present, found %d", len(hiveDataList.Items))).NotTo(HaveOccurred())
			}
		})

		It("Should delete hivePolicy after deletion of HivePolicy", func() {

			By("Deleting the HivePolicy")
			err = Client.Delete(ctx, hiveTestPolicy)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(reconcileTimeout)
		})
	})
})
