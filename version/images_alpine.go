// +build alpine,!arm

package version

import (
	"fmt"
)

const DOCKER_OS = "alpine"

var (
	ERIS_REG_DEF = "quay.io"
	ERIS_REG_BAK = "" //dockerhub

	ERIS_IMG_BASE = fmt.Sprintf("eris/base:%s", DOCKER_OS)
	ERIS_IMG_DATA = fmt.Sprintf("eris/data:%s", DOCKER_OS)
	ERIS_IMG_KEYS = fmt.Sprintf("eris/keys:%s", DOCKER_OS)
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s-%s", DOCKER_OS, VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s-%s", DOCKER_OS, VERSION)
	ERIS_IMG_CM   = fmt.Sprintf("eris/eris-cm:%s-%s", DOCKER_OS, VERSION)
	ERIS_IMG_IPFS = fmt.Sprintf("eris/ipfs:%s", DOCKER_OS)
)
