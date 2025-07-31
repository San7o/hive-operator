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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Namespace for testing
const testNamespaceName = "hive-test"

// Namespace used by the operator
const operatorNamespace = "hive-operator-system"

// Maximum time for the operator to reconcile ruccessfully
const reconcileTimeout = 2 * time.Second

// Maximum time spent waiting for creation / deletion of pods
const timeout = 30 * time.Second

var (
	Client          client.Client
	ctx             context.Context
	InitialHiveData int
)

var testNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: testNamespaceName,
	},
}
