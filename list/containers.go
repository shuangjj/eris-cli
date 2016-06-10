package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-logger"
	docker "github.com/fsouza/go-dockerclient"
)

const (
	// `eris ls` format.
	standardTmplHeader = "{{toupper .}}\tON\tCONTAINER ID\tDATA CONTAINER"
	standardTmpl       = "{{.ShortName}}\t{{asterisk .Info.State.Running}}\t{{short .Info.ID}}\t{{short (dependent .ShortName)}}"

	// `eris ls -a` format.
	extendedTmplHeader = "{{toupper .}}\tON\tCONTAINER ID\tDATA CONTAINER\tIMAGE\tCOMMAND\tPORTS"
	extendedTmpl       = "{{.ShortName}}\t{{asterisk .Info.State.Running}}\t{{short .Info.ID}}\t{{short (dependent .ShortName)}}\t{{.Info.Config.Image}}\t{{.Info.Config.Cmd}}\t{{ports .Info}}"

	// Data section.
	dataTmplHeader = "{{toupper .}}\tON\tCONTAINER ID"
	dataTmpl       = "{{.ShortName}}\t{{asterisk .Info.State.Running}}\t{{short .Info.ID}}"
)

var (
	erisContainers = []*util.Details{}

	// Template helpers to manipulate raw field values in the output.
	helpers = map[string]interface{}{
		"toupper": func(word string) string {
			return strings.ToUpper(word)
		},
		"quote": func(word string) string {
			return strconv.Quote(word)
		},
		// Show a '*' symbol if a container is running.
		"asterisk": func(running bool) string {
			if running {
				return "*"
			}
			return "-"
		},
		// Truncate the longer ID version (handy for copying and pasting).
		"short": func(id string) string {
			if len(id) <= 10 {
				return id
			}

			return id[:10]
		},
		// Show a dependent data container name if it exists
		// for the given short name of a service or a chain.
		"dependent": func(name string) string {
			for _, container := range erisContainers {
				if container.ShortName == name && container.Type == def.TypeData {
					return container.Info.ID
				}
			}
			return ""
		},
		// Pretty-format Docker ports.
		"ports": func(container *docker.Container) string {
			return util.FormulatePortsOutput(container)
		},
	}
)

// Containers display container information on the console in a format
// specified by the "format" parameter: the default "" and "extended" use the
// predefined Go templates, "json" dumps the JSON document of container
// details for every container. A custom format can be specified using
// the Go template syntax.
func Containers(t, format string, running bool) error {
	log.WithFields(log.Fields{
		"format": format,
		"type":   t,
	}).Debug("Listing containers")

	// Dump a JSON document then terminate.
	if format == "json" {
		return jsonContainers(t, running)
	}

	// Collect container information.
	util.ErisContainers(func(name string, details *util.Details) bool {
		if running == true && details.Info.State.Running == false && details.Type != def.TypeData {
			return false
		}
		erisContainers = append(erisContainers, details)
		return true
	}, false)

	// Keys for the parameter map.
	const (
		Standard = iota
		Extended
		Custom
	)
	key := Standard
	switch {
	case format == "extended":
		key = Extended
	case format != "":
		key = Custom
	}

	// Use a table to select template rendering parameters to avoid multiple nested ifs.
	buf := new(bytes.Buffer)
	renderParams := map[string]map[int][]struct {
		Type         string
		DontShowData bool
		Header       string
		Template     string
	}{
		def.TypeService: {
			Standard: {{t, false, standardTmplHeader, standardTmpl}},
			Extended: {{t, false, extendedTmplHeader, extendedTmpl}},
			Custom:   {{t, false, "", format}},
		},
		def.TypeChain: {
			Standard: {{t, false, standardTmplHeader, standardTmpl}},
			Extended: {{t, false, extendedTmplHeader, extendedTmpl}},
			Custom:   {{t, false, "", format}},
		},
		def.TypeData: {
			Standard: {{t, false, dataTmplHeader, dataTmpl}},
			Extended: {{t, false, dataTmplHeader, dataTmpl}},
			Custom:   {{t, false, "", format}},
		},
		"all": {
			Standard: {
				{def.TypeService, false, standardTmplHeader, standardTmpl},
				{def.TypeChain, false, standardTmplHeader, standardTmpl},
				{def.TypeData, true, dataTmplHeader, dataTmpl},
			},
			Extended: {
				{def.TypeService, false, extendedTmplHeader, extendedTmpl},
				{def.TypeChain, false, extendedTmplHeader, extendedTmpl},
				{def.TypeData, true, dataTmplHeader, dataTmpl},
			},
			Custom: {
				{def.TypeService, false, "", format},
				{def.TypeChain, false, "", format},
				{def.TypeData, false, "", format},
			},
		},
	}

	if _, ok := renderParams[t]; !ok {
		return fmt.Errorf("Don't know the type %q to list containers for", t)
	}

	for _, p := range renderParams[t][key] {
		// Skip the Data section altogether if there's nothing to show.
		if p.DontShowData == true && isOrphanDataContainers() == false {
			continue
		}
		if err := render(buf, p.Type, p.DontShowData, p.Header, p.Template); err != nil {
			return err
		}
	}

	// 6 - minwidth, 1 - tabwidth (tab characters width), 5 - padding, ' ' - padchar, 0 - flags.
	tw := tabwriter.NewWriter(os.Stdout, 6, 1, 5, ' ', 0)
	buf.WriteTo(tw)
	tw.Flush()

	return nil
}

func isOrphanDataContainers() bool {
	for _, container := range erisContainers {
		if container.Type == def.TypeData {
			if isMasterContainer(container.ShortName) {
				continue
			}
			return true
		}
	}
	return false
}

func isMasterContainer(name string) bool {
	// Found chain.
	if _, err := util.Lookup(def.TypeChain, name); err == nil {
		return true
	}
	// Found service.
	if _, err := util.Lookup(def.TypeService, name); err == nil {
		return true
	}
	return false
}

func render(buf *bytes.Buffer, t string, truncate bool, header, format string) error {
	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	if header != "" {
		tmplHeader, err := template.New("header").Funcs(helpers).Parse(r.Replace(header))
		if err != nil {
			return fmt.Errorf("Header template error: %v", err)
		}
		if err := tmplHeader.Execute(buf, t); err != nil {
			return fmt.Errorf("Header template exec error: %v", err)
		}
		buf.WriteString("\n")
	}

	tmplTable, err := template.New("containers").Funcs(helpers).Parse(r.Replace(format))
	if err != nil {
		return fmt.Errorf("Listing template error: %v", err)
	}

	for _, container := range erisContainers {
		// Show containers for this section type.
		if container.Type != t {
			continue
		}

		// Display only orphaned data containers in `eris ls` or `eris ls -a` mode.
		if truncate {
			if isMasterContainer(container.ShortName) {
				continue
			}
		}

		if err := tmplTable.Execute(buf, container); err != nil {
			return fmt.Errorf("Listing template exec error: %v\n", err)
		}

		buf.WriteString("\n")
	}

	if header != "" {
		// Tabs are necessary so that the Tabwriter doesn't break
		// on a newline (1 tab per column).
		buf.WriteString("\t\t\t\t\t\t\n")
	}
	return nil
}

func jsonContainers(t string, running bool) error {
	// Collect container information.
	util.ErisContainers(func(name string, details *util.Details) bool {
		if t == "all" || t == details.Type {
			erisContainers = append(erisContainers, details)
		}
		return true
	}, running)

	b, err := json.Marshal(erisContainers)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	out.WriteTo(os.Stdout)
	io.WriteString(os.Stdout, "\n")

	return nil
}
