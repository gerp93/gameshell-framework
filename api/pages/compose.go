package apiPages

import (
	"html/template"
	"io/fs"

	"github.com/gerp93/gameshell-framework/static"
)

// ParseGameFragment composes a framework-owned chrome page (base.html plus
// one named body file, e.g. "html/pages/body/deck-detail-chrome.html") with
// a game-supplied fragment file that fills the named blocks the chrome
// references (e.g. {{block "card-management" .}}{{end}}).
//
// chromeBodyName must be the only file in this call defining "body" — every
// page body template in this framework and its consuming games shares that
// one template name, and text/template silently lets a later Parse/ParseFS
// call overwrite an earlier definition of the same name. Do not pass a
// gamePatterns entry that also defines "body" (a game's own page body file,
// for instance): the second one parsed would silently replace the chrome,
// with no compile-time signal. Game fragment files must define distinctly
// named blocks instead — "card-management", "card-search-controls", etc.
func ParseGameFragment(gameFS fs.FS, chromeBodyName string, gamePatterns ...string) (*template.Template, error) {
	t, err := template.New("base.html").ParseFS(static.StaticFiles, "html/pages/base.html", chromeBodyName)
	if err != nil {
		return nil, err
	}
	return t.ParseFS(gameFS, gamePatterns...)
}
