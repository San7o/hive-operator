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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Namespace for testing
const testNamespaceName = "kive-test"

// Namespace used by the operator
const operatorNamespace = "kivebpf-system"

// Maximum time for the operator to reconcile ruccessfully
const reconcileTimeout = 2 * time.Second

// Maximum time spent waiting for creation / deletion of pods
const timeout = 30 * time.Second

var (
	Client          client.Client
	ctx             context.Context
	InitialKiveData int
)

var testNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: testNamespaceName,
	},
}
