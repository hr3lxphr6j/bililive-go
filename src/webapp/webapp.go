//go:build release

package webapp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed build
var buildFS embed.FS

func FS() (http.FileSystem, error) {
	ret, err := fs.Sub(buildFS, "build")
	if err != nil {
		return nil, err
	}
	return http.FS(ret), nil
}
