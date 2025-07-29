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
	"context"
	"fmt"
	"time"
	
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	
	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

// Namespace for testing
const namespaceName = "hive-test"
// Namespace used by the operator
const operatorNamespace = "hive-operator-system"
// Maximum time for the operator to reconcile ruccessfully
const reconcileTimeout = 1 * time.Second
// Maximum time spent waiting for creation / deletion of pods
const timeout = 30 * time.Second

var (
	Client client.Client
	ctx    context.Context
	InitialHiveData int
)

var testNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: namespaceName,
	},
}

var	hiveTestPolicy = &hivev1alpha1.HivePolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name: "hive-policy-test1",
		Namespace: namespaceName,
	},
	Spec: hivev1alpha1.HivePolicySpec{
		Path: "/test",
		Create: true,
		Match: hivev1alpha1.HivePolicyMatch{
			PodName: "test-pod",
			Namespace: "hive-test",
		},
	},
}
		
var	testPod = &corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-pod",
		Namespace: namespaceName,
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{{
			Name: "test-pod",
			Image: "nginx:latest",
			Ports: []corev1.ContainerPort{{
				ContainerPort: 80,
			}},
		}},
	},
}

var _ = Describe("Hive operator", Ordered, func() {
	var err error

	BeforeAll(func() {
		ctx = context.Background()
		Client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())

		err = CreateTestNamespace(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = CleanHivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = CleanTestPods(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		err = CleanTestPods(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = CleanHivePolicies(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
		err = DeleteTestNamespace(ctx, Client)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Operator", func() {
		
		It("Should not have any HivePolicy", func() {

			By("Getting HivePolicy")
			var hivePolicyList hivev1alpha1.HivePolicyList
			err := Client.List(ctx, &hivePolicyList, client.InNamespace(namespaceName))
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
			err := Client.List(ctx, &hivePolicyList, client.InNamespace(namespaceName))
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
			err = Client.Create(ctx, testPod)
			if err != nil {
				Expect(fmt.Errorf("Creating Test Pod: %w", err)).NotTo(HaveOccurred())
			}

			By("Waiting for pod cration")
			key := client.ObjectKeyFromObject(testPod)
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
				Expect(fmt.Errorf("HiveData should be present")).NotTo(HaveOccurred())
			}
		})
	})
})

/*
var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {
		By("installing prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		By("uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			// projectimage stores the name of the image used in the example
			var projectimage = "example.com/hive-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(projectimage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

		})
	})
})
*/
