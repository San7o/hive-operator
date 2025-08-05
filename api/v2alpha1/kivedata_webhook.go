/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package v2alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kivedatalog = logf.Log.WithName("kivedata-resource")

func (r *KiveData) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (r *KiveData) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *KiveData) ValidateUpdate(ctx context.Context, _ runtime.Object, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *KiveData) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// For mutation webhook
func (r *KiveData) SetupMutateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r). // Mutation logic
		WithCustomPath(HiveDataMutateWebhookPath).
		Complete()
}

func (r *KiveData) SetupValidateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r). // Validation logic
		WithCustomPath(HiveDataValidateWebhookPath).
		Complete()
}
