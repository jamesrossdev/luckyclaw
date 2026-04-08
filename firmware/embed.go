package firmware

import "embed"

//go:embed overlay/etc/init.d/S99luckyclaw overlay/etc/profile.d/luckyclaw-banner.sh
var FS embed.FS
