package tls

import (
	"embed"
)

//go:embed "cert"
var SSLCertificats embed.FS

// ## The important line here is //go:embed "html" "static" .
// general format go:embed <paths>
// So in our case, go:embed "static" "html"
// embeds the directories ui/static and ui/html from our project.
