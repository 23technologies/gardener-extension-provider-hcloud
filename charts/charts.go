package charts

import (
	"embed"
)

// InternalChart embeds the internal charts in embed.FS
//
//go:embed all:internal
var InternalChart embed.FS

const InternalChartsPath = "internal"
