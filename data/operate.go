package data

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

//var serv *definitions.Service

func ImportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		//need version to pick API or hack below
		dVer, err := util.DockerClientVersion()
		if err != nil {
			return err
		}

		//need service.ID for PutContainerArchive()
		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}

		containerName := util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		importPath := filepath.Join(DataContainersPath, do.Name)

		// temp until docker cp works both ways.
		logger.Debugf("Importing FROM =>\t\t%s\n", importPath)
		os.Chdir(importPath)
		// TODO [eb]: deal with hardcoded user
		// TODO [csk]: drop the whole damn cmd call
		//         use go's tar lib to make a tarball of the directory
		//         read the tar file into an io.Reader
		//         start a container with its Stdin open, connect to an io.Writer
		//         connect them up with io.Pipe
		//         this will free us from any quirks that the cli has

		// do.Path is the destination
		// if nothing is given we assume
		if do.Path == "" {
			do.Path = ErisContainerRoot
		}

		logger.Debugf("Importing TO =>\t\t\t%s\n", do.Path)
		if dVer >= 1.8 {
			reader, err := util.Tar(importPath, 0)
			if err != nil {
				return err
			}
			defer reader.Close()

			opts := docker.UploadToContainerOptions{
				InputStream:          reader,
				Path:                 do.Path,
				NoOverwriteDirNonDir: false,
			}

			logger.Infof("Copying into Cont. ID =>\t%s\n", service.ID)
			logger.Debugf("\tPath =>\t\t\t%s\n", do.Path)
			if err := util.DockerClient.UploadToContainer(service.ID, opts); err != nil {
				return err
			}

			chown := []string{"chown", "--recursive", "eris:", do.Path}
			_, err = perform.DockerRunVolumesFromContainer(containerName, false, chown, nil)
			if err != nil {
				return fmt.Errorf("fack: %v\n", err)
			}
		} else {
			cmd := "tar chf - . | docker run -i --rm --volumes-from " + containerName + " --user eris eris/data tar xf - -C " + do.Path
			_, err := pipes.RunString(cmd)
			if err != nil {
				cmd := "tar chf - . | docker run -i --volumes-from " + containerName + " --user eris eris/data tar xf - -C " + do.Path
				_, e2 := pipes.RunString(cmd)
				if e2 != nil {
					return fmt.Errorf("Could not import the data container.\nTried with docker --rm =>\t%v\nTried without docker --rm =>\t%v", err, e2)
				}
			}
		}
	} else {
		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		if err := perform.DockerCreateDataContainer(ops); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}
		return ImportData(do)
	}
	do.Result = "success"
	return nil
}

func ExecData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		logger.Infoln("Running exec on container with volumes from data container " + do.Operations.DataContainerName)

		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		util.Merge(ops, do.Operations)
		if _, err := perform.DockerRunVolumesFromContainer(ops, nil); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("The marmots cannot find that data container.\nPlease check the name of the data container with [eris data ls].")
	}
	do.Result = "success"
	return nil
}

func ExportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		dVer, err := util.DockerClientVersion()
		if err != nil {
			return err
		}

		logger.Infoln("Exporting data container", do.Name)

		// we want to export to a temp directory.
		exportPath, err := ioutil.TempDir(os.TempDir(), do.Name) // TODO: do.Operations.ContainerNumber ?
		defer os.Remove(exportPath)
		if err != nil {
			return err
		}

		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}

		cont, err := util.DockerClient.InspectContainer(service.ID)
		if err != nil {
			return err
		}

		reader, writer := io.Pipe()
		defer reader.Close()

		if do.Path != "" {
			do.Path = do.Path
		} else {
			do.Path = "/home/eris/.eris"
		}
		opts := docker.DownloadFromContainerOptions{
			OutputStream: writer,
			//	Container:    service.ID,
			Path: do.Path,
		}

		go func() {
			logger.Infof("Copying out of Cont. ID =>\t%s\n", service.ID)
			logger.Debugf("\tPath =>\t\t\t%s\n", do.Path)
			IfExit(util.DockerClient.DownloadFromContainer(service.ID, opts)) // TODO: be smarter about catching this error
			writer.Close()
		}()

		logger.Debugf("Untarring Package from Cont =>\t%s\n", exportPath)
		if err = util.Untar(reader, do.Name, exportPath); err != nil {
			return err
		}

		// docker actually exports to a `_data` folder for volumes
		//   this section of the function moves whatever docker dumps
		//   into exportPath/_data into export. ranging through the
		//   volumes is probably overkill as we could just assume
		//   that it *was* `_data` but in case docker changes later
		//   we'll just keep it for now. this is specific to 1.7 and
		//   below. For 1.8 we do not need to do this.
		// os.Chdir(exportPath)
		if dVer <= 1.7 {
			var unTarDestination string
			logger.Debugf("Container's Volumes =>\t\t%v\n", cont.Volumes)
			for k, v := range cont.Volumes {
				if k == do.Path {
					unTarDestination = filepath.Base(v)
				}
			}
			logger.Debugf("Untarring to =>\t\t\t%s:%s\n", exportPath, unTarDestination)
			if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, unTarDestination), exportPath); err != nil {
				return err
			}
		} else {
			logger.Debugf("Untarring to =>\t\t\t%s\n", exportPath)
		}

		// now if docker dumps to exportPath/.eris we should remove
		//   move everything from .eris to exportPath
		if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, ".eris"), exportPath); err != nil {
			return err
		}

		// // finally remove everything in the data directory and move
		// //   the temp contents there
		prevDir := filepath.Join(DataContainersPath, do.Name)
		if _, err := os.Stat(prevDir); os.IsNotExist(err) {
			if e2 := os.MkdirAll(prevDir, 0755); e2 != nil {
				return fmt.Errorf("Error:\tThe marmots could neither find, nor had access to make the directory: (%s)\n", prevDir)
			}
		}
		// ClearDir(prevDir)
		if err := moveOutOfDirAndRmDir(exportPath, prevDir); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	do.Result = "success"
	return nil
}

func moveOutOfDirAndRmDir(src, dest string) error {
	logger.Infof("Move all files/dirs out of a dir and rm -rf that dir.\n")
	logger.Debugf("Source of the move =>\t\t%s.\n", src)
	logger.Debugf("Destin of the move =>\t\t%s.\n", dest)
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		logger.Debugln("No files to move.")
	}

	for _, f := range toMove {
		logger.Debugf("Moving [%s] to [%s].\n", f, filepath.Join(dest, filepath.Base(f)))

		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		Copy(f, filepath.Join(dest, filepath.Base(f)))
	}

	logger.Infof("Removing directory =>\t\t%s.\n", src)
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}
