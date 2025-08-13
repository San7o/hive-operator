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

	kivev2alpha1 "github.com/San7o/kivebpf/api/v2alpha1"
)

var _ = Describe("ContainerName Regex 2", Ordered, func() {
	var err error

	var kiveTestPolicy = &kivev2alpha1.KivePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kive-policy-test-regex2",
			Namespace: testNamespaceName,
		},
		Spec: kivev2alpha1.KivePolicySpec{
			Traps: []kivev2alpha1.KiveTrap{
				{
					Path:   "/regex2",
					Create: true,
					MatchAny: []kivev2alpha1.KiveTrapMatch{
						kivev2alpha1.KiveTrapMatch{
							PodName:       "test-pod",
							Namespace:     "kive-test",
							ContainerName: "regex:test-nope.*",
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
		err = CleanKivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = CleanTestPods(ctx, Client, []corev1.Pod{testPod})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		err = CleanTestPods(ctx, Client, []corev1.Pod{testPod})
		Expect(err).NotTo(HaveOccurred())
		err = CleanKivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Operator", func() {

		It("Should not have any KivePolicy", func() {

			By("Getting KivePolicy")
			var kivePolicyList kivev2alpha1.KivePolicyList
			err := Client.List(ctx, &kivePolicyList, client.InNamespace(testNamespaceName))
			Expect(err).NotTo(HaveOccurred())

			if len(kivePolicyList.Items) != 0 {
				Expect(fmt.Errorf("KivePolicy present")).NotTo(HaveOccurred())
			}
		})

		It("Should not have any KiveData", func() {

			By("Getting KiveData")
			var kiveDataList kivev2alpha1.KiveDataList
			err := Client.List(ctx, &kiveDataList, client.InNamespace(operatorNamespace))
			Expect(err).NotTo(HaveOccurred())

			if len(kiveDataList.Items) != 0 {
				Expect(fmt.Errorf("KiveData present")).NotTo(HaveOccurred())
			}
		})

		It("Should succesfully create an KivePolicy", func() {

			By("Creating KivePolicy")
			err = Client.Create(ctx, kiveTestPolicy)
			Expect(err).NotTo(HaveOccurred())

			// Give the operator some time to react
			time.Sleep(reconcileTimeout)

			By("Getting KivePolicy")
			var kivePolicyList kivev2alpha1.KivePolicyList
			err := Client.List(ctx, &kivePolicyList, client.InNamespace(testNamespaceName))
			Expect(err).NotTo(HaveOccurred())

			if len(kivePolicyList.Items) != 1 {
				Expect(fmt.Errorf("KivePolicy not present")).NotTo(HaveOccurred())
			}

			By("Getting KiveData")
			var kiveDataList kivev2alpha1.KiveDataList
			err = Client.List(ctx, &kiveDataList, client.InNamespace(operatorNamespace))
			Expect(err).NotTo(HaveOccurred())

			if len(kiveDataList.Items) != 0 {
				Expect(fmt.Errorf("KiveData should not be present")).NotTo(HaveOccurred())
			}
		})

		It("Should not create an KiveData when a pod does not match the container name", func() {

			By("Creating test pod")
			err = Client.Create(ctx, &testPod)
			if err != nil {
				Expect(fmt.Errorf("Creating Test Pod: %w", err)).NotTo(HaveOccurred())
			}

			By("Waiting for pod cration")
			key := client.ObjectKeyFromObject(&testPod)
			deadline := time.Now().UTC().Add(timeout)
			for time.Now().UTC().Before(deadline) {
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

			By("Getting KiveData")
			var kiveDataList kivev2alpha1.KiveDataList
			if err := Client.List(ctx, &kiveDataList, client.InNamespace(operatorNamespace)); err != nil {
				Expect(fmt.Errorf("List KiveData: %w", err)).NotTo(HaveOccurred())
			}
			if len(kiveDataList.Items) != 0 {
				Expect(fmt.Errorf("One KiveData should not be present, found %d", len(kiveDataList.Items))).NotTo(HaveOccurred())
			}
		})

		It("Should delete kivePolicy after deletion of KivePolicy", func() {

			By("Deleting the KivePolicy")
			err = Client.Delete(ctx, kiveTestPolicy)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(reconcileTimeout)
		})
	})
})
