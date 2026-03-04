package data

import "embed"

//go:embed bin/static.7z
var StaticZipFile embed.FS
