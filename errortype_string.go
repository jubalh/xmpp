// Code generated by "stringer -type=errorType"; DO NOT EDIT

package xmpp

import "fmt"

const _errorType_name = "AuthCancelContinueModifyWait"

var _errorType_index = [...]uint8{0, 4, 10, 18, 24, 28}

func (i errorType) String() string {
	if i < 0 || i >= errorType(len(_errorType_index)-1) {
		return fmt.Sprintf("errorType(%d)", i)
	}
	return _errorType_name[_errorType_index[i]:_errorType_index[i+1]]
}
