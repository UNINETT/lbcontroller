// Code generated by "stringer -type action"; DO NOT EDIT.

package nlb

import "strconv"

const _action_name = "replacereconfigdelete"

var _action_index = [...]uint8{0, 7, 15, 21}

func (i action) String() string {
	if i < 0 || i >= action(len(_action_index)-1) {
		return "action(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _action_name[_action_index[i]:_action_index[i+1]]
}