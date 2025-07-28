package ebpf

import (
	"fmt"

	"github.com/cilium/ebpf"
)

/*
 *  Update element at index with value in the TracedInodes map.
 */
func UpdateTracedInodes(index uint32, value uint64) error {

	err := Objs.TracedInodes.Update(index, value, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("UpdateTracedInodes Error: %w", err)
	}

	return nil
}

/*
 * Fills a map with zeroes from index to MapMaxEntries.
 */
func ResetTracedInodes(index uint32) error {

	for ; index < MapMaxEntries; index++ {
		err := UpdateTracedInodes(index, uint64(0))
		if err != nil {
			return fmt.Errorf("ResetMap Error: %w", err)
		}
	}

	return nil
}
