package version

import (
	"fmt"
)

var (
	ERIS_REG_DEF = "" //quay.io
	ERIS_REG_BAK = "" //dockerhub

	ERIS_IMG_BASE = fmt.Sprintf("eris4iot/base:%s", ARCH_ARM)
	ERIS_IMG_DATA = fmt.Sprintf("eris4iot/data:%s", ARCH_ARM)
	ERIS_IMG_KEYS = fmt.Sprintf("eris4iot/keys:%s", ARCH_ARM)
	ERIS_IMG_DB   = fmt.Sprintf("eris4iot/erisdb:%s", ARCH_ARM)
	ERIS_IMG_PM   = fmt.Sprintf("eris4iot/epm:%s", ARCH_ARM)
	ERIS_IMG_CM   = fmt.Sprintf("eris4iot/eris-cm:%s", ARCH_ARM)
	ERIS_IMG_IPFS = fmt.Sprintf("eris4iot/ipfs:%s", ARCH_ARM)
)
