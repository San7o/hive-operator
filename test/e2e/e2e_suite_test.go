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
	"context"
	"fmt"
	"testing"

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
