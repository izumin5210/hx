package hx

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func Path(elem ...interface{}) string {
	chunks := make([]string, len(elem))
	for i, e := range elem {
		var s string
		switch v := e.(type) {
		case string:
			s = v
		case fmt.Stringer:
			s = v.String()
		default:
			s = fmt.Sprint(v)
		}
		chunks[i] = s
	}
	if u, err := url.Parse(chunks[0]); err == nil && u.IsAbs() {
		return strings.TrimSuffix(chunks[0], "/") + "/" + path.Join(chunks[1:]...)
	}
	return path.Join(chunks...)
}
