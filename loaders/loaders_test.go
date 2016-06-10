package loaders

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/eris-ltd/common/go/common"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-logger"
)

type ab struct {
	name string
	a, b interface{}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.DontPull))

	exitCode := m.Run()

	tests.IfExit(tests.TestsTearDown())

	os.Exit(exitCode)
}

func TestLoadChainDefinitionEmptyDefault(t *testing.T) {
	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "1234" ]
`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.ChainID, name},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionEmptyDefinition(t *testing.T) {
	const (
		name = "test"

		defaultDefinition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "1234" ]

[dependencies]
services = [ "keys" ]

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"
`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", defaultDefinition); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.ChainID, ""},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		// [pv]: "data_container" is not loaded from the default.toml. A bug?
		{`Service.AutoData`, d.Service.AutoData, false},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},

		{`Dependencies`, d.Dependencies.Services, []string{"keys"}},
		{`Maintainer`, d.Maintainer.Email, "support@erisindustries.com"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionEmptyDefaultAndDefinition(t *testing.T) {
	const (
		name = "test"
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.ChainID, ""},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionOverwrite(t *testing.T) {
	const (
		name = "test"

		defaultDefinition = `
[service]
name           = "random default name"
image          = "default image"
data_container = true
ports          = [ "1234" ]

[dependencies]
chains = [ "something" ]

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"
`

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "4321" ]
`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", defaultDefinition); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.ChainID, name},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		// [pv]: ports are mixed, not overwritten! (util.Merge behaviour)
		{`Service.Ports`, d.Service.Ports, []string{"1234", "4321"}},

		{`Dependencies`, d.Dependencies.Chains, []string{"something"}},
		{`Maintainer`, d.Maintainer.Email, "support@erisindustries.com"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionMissingDefault(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ChainsPath, "default.toml"))

	if err := tests.FakeDefinitionFile(common.ChainsPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadChainDefinition(name); err == nil {
		t.Fatalf("expected load to fail")
	}
}

func TestLoadChainDefinitionBadFormatDefault(t *testing.T) {
	const (
		name = "test"

		defaultDefinition = `name = [ "keys" ]`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", defaultDefinition); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadChainDefinition(name); err == nil {
		t.Fatalf("expected load to fail")
	}
}

func TestLoadChainDefinitionMissingDefinition(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ChainsPath, name+".toml"))

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}

	if _, err := LoadChainDefinition(name); err == nil {
		t.Fatalf("expected load to fail")
	}
}

func TestLoadChainDefinitionBadFormatDefinition(t *testing.T) {
	const (
		name = "test"

		definition = `name = [ "keys" ]`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadChainDefinition(name); err == nil {
		t.Fatalf("expected load to fail")
	}
}

func TestChainsAsAServiceSimple(t *testing.T) {
	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
data_container = true
ports          = [ "1234" ]
image          = "test image"
`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(common.ChainsPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	s, err := ChainsAsAService(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, s.Name, name},
		{`ContainerType`, s.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, s.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, s.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, s.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, s.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, s.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, s.Service.Name, name},
		{`Service.AutoData`, s.Service.AutoData, true},
		// [pv]: not "test image", but erisdb image. A bug?
		{`Service.Image`, s.Service.Image, path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_DB)},
		{`Service.Environment`, s.Service.Environment, []string{"CHAIN_ID=" + name}},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestChainsAsAServiceMissing(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ChainsPath, name+".toml"))

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}

	if _, err := ChainsAsAService(name); err == nil {
		t.Fatalf("expected chains as a service to fail")
	}
}

func TestMockChainDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := MockChainDefinition(name, "id")

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.ChainID, "id"},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("mock definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}

}

func TestMockChainDefinitionBlankNames(t *testing.T) {
	d := MockChainDefinition("", "")

	for _, entry := range []ab{
		{`Name`, d.Name, ""},
		{`ChainID`, d.ChainID, ""},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName("")},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName("")},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], ""},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("mock definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadDataDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := LoadDataDefinition(name)

	for _, entry := range []ab{
		{`ContainerType`, d.ContainerType, def.TypeData},
		{`SrvContainerName`, d.SrvContainerName, util.DataContainerName(name)},
		{`DataContainerName`, d.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Labels[def.LabelType], def.TypeData},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageSimple(t *testing.T) {
	const (
		name = "test"

		definition = `
[eris]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(common.ErisRoot, name)
	if err != nil {
		t.Fatalf("expected to load definition file, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`PackageID`, d.PackageID, name},
		{`ChainName`, d.ChainName, "test chain"},
		{`ChainID`, d.ChainID, "test id"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageDirectoryAndSpacesInAName(t *testing.T) {
	const (
		name = "test test"

		definition = `
name       = "` + name + `"

[eris]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(filepath.Join(common.ErisRoot, "package.toml"), name)
	if err != nil {
		t.Fatalf("expected to load definition file, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "test_test"},
		{`PackageID`, d.PackageID, name},
		{`ChainName`, d.ChainName, "test chain"},
		{`ChainID`, d.ChainID, "test id"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageNotFound1(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	if _, err := LoadPackage("/non/existent/path", name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadPackageNotFound2(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	d, err := LoadPackage(common.ErisRoot, "")
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "eris"},
		{`PackageID`, d.PackageID, ""},
		{`ChainName`, d.ChainName, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageNotFound3(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	d, err := LoadPackage(common.ErisRoot, name)
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "eris"},
		{`PackageID`, d.PackageID, ""},
		{`ChainName`, d.ChainName, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageBadFormat(t *testing.T) {
	const (
		name = "test"

		definition = `
[eris]
name       = [ "keys"]
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadPackage(common.ErisRoot, name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionSimple(t *testing.T) {
	const (
		name       = "test"
		definition = `
name = "` + name + `"
description = "description"
status = "in production"

[service]
image = "test image"
data_container = true
ports = [ "1234" ]

[location]
repository = "https://example.com"
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},

		{`Location`, d.Location.Repository, "https://example.com"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadServiceDefinitionAlmostEmpty(t *testing.T) {
	const (
		name       = "test"
		definition = `
[service]
image = "test image"
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "test image"},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName("test image")},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName("test image")},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, "test image"},
		{`Service.Image`, d.Service.Image, "test image"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadServiceDefinitionEmpty(t *testing.T) {
	const (
		name = "test"
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionMissing(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ServicesPath, name+".toml"))

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionBadFormat(t *testing.T) {
	const (
		name = "test"

		definition = `
[service]
image = [ "keys" ]
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestMockServiceDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := MockServiceDefinition(name)

	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, name},
		// [pv]: Mock is allowed to return an empty image while load isn't?
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankNames(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Service.Image = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankName(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Service.Name = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankServiceName(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Name = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankAllTheThings(t *testing.T) {
	defer func() {
		recover()
	}()

	d := def.BlankServiceDefinition()

	ServiceFinalizeLoad(d)

	t.Fatalf("expected finalize to panic")
}
