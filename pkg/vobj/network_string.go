// Code generated by "stringer -type=Network -linecomment"; DO NOT EDIT.

package vobj

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NetworkUnknown-0]
	_ = x[NetworkHttp-1]
	_ = x[NetworkHttps-2]
	_ = x[NetworkRpc-3]
}

const _Network_name = "未知httphttpsrpc"

var _Network_index = [...]uint8{0, 6, 10, 15, 18}

func (i Network) String() string {
	if i < 0 || i >= Network(len(_Network_index)-1) {
		return "Network(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Network_name[_Network_index[i]:_Network_index[i+1]]
}

// IsUnknown 是否是：未知
func (i Network) IsUnknown() bool {
	return i == NetworkUnknown
}

// IsHttp 是否是：http
func (i Network) IsHttp() bool {
	return i == NetworkHttp
}

// IsHttps 是否是：https
func (i Network) IsHttps() bool {
	return i == NetworkHttps
}

// IsRpc 是否是：rpc
func (i Network) IsRpc() bool {
	return i == NetworkRpc
}

// GetValue 获取原始类型值
func (i Network) GetValue() int {
	return int(i)
}
