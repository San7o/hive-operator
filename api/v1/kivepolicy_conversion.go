/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package v1

import (
	v2alpha1 "github.com/San7o/kivebpf/api/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// From v1 to v2alpha1
func (src *KivePolicy) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v2alpha1.KivePolicy)

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()

	trapsv2 := []v2alpha1.KiveTrap{}
	for _, trap := range src.Spec.Traps {

		trapv2 := v2alpha1.KiveTrap{}
		trapv2.Path = trap.Path
		trapv2.Create = trap.Create
		trapv2.Mode = trap.Mode
		trapv2.Callback = trap.Callback

		matchesv2 := []v2alpha1.KiveTrapMatch{}

		for _, match := range trap.MatchAny {

			matchv2 := v2alpha1.KiveTrapMatch{}
			matchv2.PodName = match.PodName
			matchv2.ContainerName = match.ContainerName
			matchv2.Namespace = match.Namespace
			matchv2.IP = match.IP
			matchv2.MatchLabels = match.MatchLabels

			matchesv2 = append(matchesv2, matchv2)
		}

		trapv2.MatchAny = matchesv2
		trapsv2 = append(trapsv2, trapv2)
	}

	dst.Spec.Traps = trapsv2

	return nil
}

// From v2alpha1 to v1
func (dst *KivePolicy) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v2alpha1.KivePolicy)

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()

	trapsv1 := []KiveTrap{}
	for _, trap := range src.Spec.Traps {

		trapv1 := KiveTrap{}
		trapv1.Path = trap.Path
		trapv1.Create = trap.Create
		trapv1.Mode = trap.Mode
		trapv1.Callback = trap.Callback

		matchesv1 := []KiveTrapMatch{}

		for _, match := range trap.MatchAny {

			matchv1 := KiveTrapMatch{}
			matchv1.PodName = match.PodName
			matchv1.ContainerName = match.ContainerName
			matchv1.Namespace = match.Namespace
			matchv1.IP = match.IP
			matchv1.MatchLabels = match.MatchLabels

			matchesv1 = append(matchesv1, matchv1)
		}

		trapv1.MatchAny = matchesv1
		trapsv1 = append(trapsv1, trapv1)
	}

	dst.Spec.Traps = trapsv1

	return nil
}
