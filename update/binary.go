package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/perform"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func BuildErisBinContainer(branch, binaryPath string) error {
	// base built locally from quay.io/eris/base because parsing error...?
	base := path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_BASE)
	dockerfile := `FROM ` + base + `
MAINTAINER Eris Industries <support@erisindustries.com>

ENV NAME         eris-cli
ENV REPO 	 eris-ltd/$NAME
ENV BRANCH       ` + branch + `
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO

RUN mkdir --parents $CLONE_PATH

RUN git clone -q https://github.com/$REPO $CLONE_PATH
RUN cd $CLONE_PATH && git fetch -a origin &&  git checkout -q $BRANCH
RUN cd $CLONE_PATH/cmd/eris && go build -o $INSTALL_BASE/eris

CMD ["/bin/bash"]`

	imageName := "eris/update:temp"
	//log.Debug(dockerfile)
	//log.Debug(imageName)
	if err := perform.DockerBuild(imageName, dockerfile); err != nil {
		return err
	}

	doNew := definitions.NowDo()
	doNew.Name = "update"
	doNew.Operations.Args  = []string{imageName}
	if err := services.NewService(doNew); err != nil {
		return err
	}

	doUpdate := definitions.NowDo()
	doUpdate.Operations.Args = []string{"update"}

	if err := services.StartService(doUpdate); err != nil {
		return nil
	}

	doCp := definitions.NowDo()
	doCp.Name = "update"

	//$INSTALL_BASE/eris
	doCp.Source = "/usr/local/bin/eris"
	doCp.Destination = common.ScratchPath
	if err := data.ExportData(doCp); err != nil {
		return err
	}
	// XXX move bin from scratch to binaryPath

	doRm := definitions.NowDo()
	doRm.Operations.Args = []string{"update"}
	doRm.RmD = true
	doRm.Volumes = true
	doRm.Force = true
	doRm.File = true

	if err := services.RmService(doRm); err != nil {
		return err
	}

	//TODO remove imageName the image
	if err := perform.DockerRemoveImage(imageName, true); err != nil {
		return err
	}

	return nil
}



// XXX code below is ~ replaced by code above. 
// left here for legacy reasons / is a complementary
// approach to the build container that we may want
// to consider ... ?
func DownloadLatestBinaryRelease(binPath string) (string, error) {

	filename, fileURL, version, err := getLatestBinaryInfo()

	erisBin, output, err := createBinaryFile(filename)

	fileResponse, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error getting file: %v\n", err)
	}
	defer fileResponse.Body.Close()

	_, err = io.Copy(output, fileResponse.Body)
	if err != nil {
		return "", fmt.Errorf("error saving file: %v\n", err)
	}

	platform := runtime.GOOS
	// this is hacky !!!
	if erisBin != "" {
		log.Println("downloaded eris binary", version, "for", platform, "to", erisBin, "\n Please manually move to", binPath)
	} else {
		log.Println("downloaded eris binary", version, "for", platform, "to", binPath)
	}

	// TODO fix this part!
	var unzip string = "tar -xvf"
	if platform != "linux" {
		unzip = "unzip"
	}
	cmd := exec.Command("bin/sh", "-c", unzip, filename)
	if err := cmd.Run(); err != nil {
		return filename, fmt.Errorf("unzipping failed: %v\n", err)
	}
	// end fix needed

	return filename, nil
}

func getLatestBinaryInfo() (string, string, string, error) {
	latestURL := "https://github.com/eris-ltd/eris-cli/releases/latest"
	resp, err := http.Get(latestURL)
	if err != nil {
		return "", "", "", fmt.Errorf("could not retrieve latest eris release at %s\nerror: %v\n", latestURL, err)
	}

	latestURL = resp.Request.URL.String()
	lastPos := strings.LastIndex(latestURL, "/")
	version := latestURL[lastPos+1:]
	platform := runtime.GOOS
	arch := runtime.GOARCH
	hostURL := "https://github.com/eris-ltd/eris-cli/releases/download/" + version + "/"
	filename := "eris_" + version[1:] + "_" + platform + "_" + arch
	fileURL := hostURL + filename

	switch platform {
	case "linux":
		filename += ".tar.gz"
	default:
		filename += ".zip"
	}

	return filename, fileURL, version, nil
}

func createBinaryFile(filename string) (string, *os.File, error) {
	var erisBin string
	output, err := os.Create(filename)
	// if we dont have permissions to create a file where eris cli exists, attempt to create file within HOME folder
	if err != nil {
		erisBin := filepath.Join(common.ScratchPath, "bin")
		if _, err = os.Stat(erisBin); os.IsNotExist(err) {
			err = os.MkdirAll(erisBin, 0755)
			if err != nil {
				log.Println("Error creating directory", erisBin, "Did not download binary. Exiting...")
				return "", nil, err
			}
		}
		err = os.Chdir(erisBin)
		if err != nil {
			log.Println("Error changing directory to", erisBin, "Did not download binary. Exiting...")
			return "", nil, err
		}
		output, err = os.Create(filename)
		if err != nil {
			log.Println("Error creating file", erisBin, "Exiting...")
			return "", nil, err
		}
	}
	defer output.Close()
	return erisBin, output, nil
}
