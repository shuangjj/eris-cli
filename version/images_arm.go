package version

import (
	"fmt"
)

const ARCH = "arm"

var (
	ERIS_REG_DEF = "quay.io"
	ERIS_REG_BAK = "" //dockerhub

	ERIS_IMG_BASE = fmt.Sprintf("eris/base:%s", ARCH)
	ERIS_IMG_DATA = fmt.Sprintf("eris/data:%s", ARCH)
	ERIS_IMG_KEYS = fmt.Sprintf("eris/keys:%s", ARCH)
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s-%s", ARCH, VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s-%s", ARCH, VERSION)
	ERIS_IMG_CM   = fmt.Sprintf("eris/eris-cm:%s-%s", ARCH, VERSION)
	ERIS_IMG_IPFS = fmt.Sprintf("eris/ipfs:%s", ARCH)
)
