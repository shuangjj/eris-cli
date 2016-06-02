// +build alpine,!arm

package version

import (
	"fmt"
)

const CONTAINER_OS = "alpine"

var (
	ERIS_REG_DEF = "quay.io"
	ERIS_REG_BAK = "" //dockerhub

	ERIS_IMG_BASE = fmt.Sprintf("eris/base:%s", CONTAINER_OS)
	ERIS_IMG_DATA = fmt.Sprintf("eris/data:%s", CONTAINER_OS)
	ERIS_IMG_KEYS = fmt.Sprintf("eris/keys:%s", CONTAINER_OS)
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s-%s", CONTAINER_OS, VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s-%s", CONTAINER_OS, VERSION)
	ERIS_IMG_CM   = fmt.Sprintf("eris/eris-cm:%s-%s", CONTAINER_OS, VERSION)
	ERIS_IMG_IPFS = fmt.Sprintf("eris/ipfs:%s", CONTAINER_OS)
)
