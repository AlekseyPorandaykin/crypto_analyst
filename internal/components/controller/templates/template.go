package templates

import (
	_ "embed"
	"time"
)

//go:embed header.html
var HeaderHtml []byte

//go:embed footer.html
var FooterHtml []byte

//go:embed index.html
var IndexHtmlPage []byte

//go:embed symbol.html
var SymbolHtmlPage []byte

//go:embed changes.html
var PriceChangesHtmlPage []byte

type PageData struct {
	Title       string
	Symbol      string
	Data        interface{}
	CurrentTime time.Time
}
