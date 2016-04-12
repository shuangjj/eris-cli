package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func UpdateEris(do *definitions.Do) error {
	// TODO organize code appropriately
	// clean up dockerfile => deal with `FROM` parsing error! or file issue...?
	// do good loggers
	// think of a test ...?
	// finish implementing / test the branch/commit/version thingy

	_, binPath, err := GoOrBinary()
	if err != nil {
		return err
	}
	// TODO check flags!

	/*if whichEris == "go" {
		hasGit, hasGo := CheckGitAndGo(true, true)
		if !hasGit || !hasGo {
			return fmt.Errorf("either git or go is not installed. both are required for non-binary update")
		}
		log.WithField("branch", do.Branch).Warn("Building eris binary via go with:")
		if err := UpdateErisGo(do); err != nil {
			return err
		}
	} else if whichEris == "binary" {*/
	log.WithField("branch", do.Branch).Warn("Building eris binary in container with:")
	if err := BuildErisBinContainer(do.Branch, binPath); err != nil {
		return err
	}
		// XXX deprecate ... ?
		//if err := UpdateErisBinary(binPath); err != nil {
		//	return err
		//}
	//} else {
	//	return fmt.Errorf("The marmots could not figure out how eris was installed")
	//}

	//checks for deprecated dir names and renames them
	// false = no prompt
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, false); err != nil {
		log.Warn(fmt.Sprintf("Directory migration error: %v\nContinuing update without migration", err))
	}
	log.Warn("Eris update successful. Please re-run `eris init`.")
	return nil
}


func GoOrBinary() (string, string, error) {
	which, err := exec.Command("which", "eris").CombinedOutput()
	if err != nil {
		return "", "", err
	}

	toCheck := strings.Split(string(which), "/")
	length := len(toCheck)
	bin := util.TrimString(toCheck[length-2])
	eris := util.TrimString(toCheck[length-1]) //sometimes ya just gotta trim

	gopath := filepath.Join(os.Getenv("GOPATH"), bin, eris)

	erisLook, err := exec.LookPath("eris")
	if err != nil {
		return "", "",  err
	}

	trimWhich := util.TrimString(string(which))
	if trimWhich != erisLook {
		return "", "", fmt.Errorf("`which eris` returned (%s) while the exec.LookPath(`eris`) command returned (%s). these need to match", trimWhich, erisLook)
	}

	if bin == "bin" && eris == "eris" {
		if util.TrimString(gopath) == util.TrimString(string(which)) { // gotta trim those strings!
			log.Debug("looks like eris was instaled via go")
			return "go", gopath, nil
		} else {
			log.Debug("looks like eris was instaled via binary")
			return "binary", erisLook, nil
		}
	} else {
		return "", "", fmt.Errorf("could not determine how eris is installed")
	}
	return "", "", err
}
