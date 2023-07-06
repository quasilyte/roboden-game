// Code generated by "stringer -type=colonyPriority -trimprefix=priority"; DO NOT EDIT.

package staging

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[priorityResources-0]
	_ = x[priorityGrowth-1]
	_ = x[priorityEvolution-2]
	_ = x[prioritySecurity-3]
}

const _colonyPriority_name = "ResourcesGrowthEvolutionSecurity"

var _colonyPriority_index = [...]uint8{0, 9, 15, 24, 32}

func (i colonyPriority) String() string {
	if i < 0 || i >= colonyPriority(len(_colonyPriority_index)-1) {
		return "colonyPriority(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _colonyPriority_name[_colonyPriority_index[i]:_colonyPriority_index[i+1]]
}
