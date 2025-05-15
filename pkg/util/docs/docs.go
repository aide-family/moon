package docs

import (
	"embed"
	nethttp "net/http"

	"github.com/go-kratos/kratos/v2/transport/http"
)

func RegisterDocs(srv *http.Server, docFS embed.FS, dev bool) {
	if !dev {
		return
	}
	doc := nethttp.FS(docFS)
	srv.HandlePrefix("/doc/", nethttp.StripPrefix("/doc/", nethttp.FileServer(doc)))
}
