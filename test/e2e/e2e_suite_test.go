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
	"testing"
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	
	// Configure Ginkgo to run sequentially
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.ParallelTotal = 1 // force sequential
	suiteConfig.RandomizeAllSpecs = false
	
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting hive-operator suite\n")
	RunSpecs(t, "Hive Operator E2E Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	var err error

	ctx = context.Background()
	Client, err = NewClient()
	Expect(err).NotTo(HaveOccurred())

	err = CreateTestNamespace(ctx, Client)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := DeleteTestNamespace(ctx, Client)
	Expect(err).NotTo(HaveOccurred())
})
