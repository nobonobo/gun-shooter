package resources

import "embed"

//go:embed ui
var UI embed.FS

//go:embed licenses.txt
var Licenses string
