//go:build windows

package readline

import "errors"

// StartEditorWithBuffer - Not implemented on Windows platforms.
func (rl *Instance) StartEditorWithBuffer(multiline []rune) ([]rune, error) {
	return rl.line, errors.New("Not currently supported on Windows")
}
