package converter

import "bytes"

// newBytesReader возвращает io.Reader из среза байт
func newBytesReader(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}
