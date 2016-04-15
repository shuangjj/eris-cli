package update

import (
	//"fmt"
	//"io"
	//"net/http"
	//"os"
	"os/exec"
	//"path/filepath"
	//"runtime"
	//"strings"

	log "github.com/Sirupsen/logrus"
	//"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		log.WithField("branch", branch).Fatalf("Error checking out branch: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch checked-out")
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error pulling from GitHub: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch pulled successfully")
}

func InstallErisGo() {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error with go install ./cmd/eris: %v", string(stdOut))
	}

	log.Debug("go install worked correctly")
}

func version() string {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("error getting version:\n%s\n", string(stdOut))
	}
	return string(stdOut)

}
