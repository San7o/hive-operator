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
var kivepolicylog = logf.Log.WithName("kiveplicy-resource")

func (r *KivePolicy) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (r *KivePolicy) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *KivePolicy) ValidateUpdate(ctx context.Context, _ runtime.Object, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *KivePolicy) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// For mutation webhook
func (r *KivePolicy) SetupMutateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r). // Mutation logic
		WithCustomPath(HivePolicyMutateWebhookPath).
		Complete()
}

func (r *KivePolicy) SetupValidateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r). // Validation logic
		WithCustomPath(HivePolicyValidateWebhookPath).
		Complete()
}
