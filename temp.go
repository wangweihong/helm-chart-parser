package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	cpb "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

const notesFileSuffix = "NOTES.txt"
const GoTplEngine = "gotpl"

func checkDependencies(ch *cpb.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}

//copied from k8s.io/helm/cmd/helm/install.go
func vals() (*cpb.Config, error) {
	base := map[string]interface{}{}
	raw, err := yaml.Marshal(base)
	if err != nil {
		return nil, err
	}
	return &cpb.Config{Raw: string(raw)}, nil

}

func main() {
	releaseName := "haha"
	chartPath := "/home/wwh/kiongf/go/src/k8s.io/helm/bin/wordpress"
	deployNamespace := "default"
	revision := 1

	chart, err := chartutil.Load(chartPath)
	if err != nil {
		panic(err.Error())
	}

	values, _ := vals()
	err = chartutil.ProcessRequirementsEnabled(chart, values)
	if err != nil {
		panic(err.Error())
	}

	err = chartutil.ProcessRequirementsImportValues(chart)
	if err != nil {
		panic(err.Error())
	}
	//处理依赖
	if req, err := chartutil.LoadRequirements(chart); err == nil {
		if err := checkDependencies(chart, req); err != nil {
			panic(err.Error())
		}

	} else if err != chartutil.ErrRequirementsNotFound {
		panic(fmt.Errorf("cannot load requirements: %v", err).Error())
	}

	ts := timeconv.Now()
	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Time:      ts,
		Namespace: deployNamespace,
		Revision:  revision,
		IsInstall: true,
	}

	valuesToRender, err := chartutil.ToRenderValues(chart, values, options)
	if err != nil {
		panic(err.Error())
	}

	e := engine.New()
	files, err := e.Render(chart, valuesToRender)
	if err != nil {
		panic(err.Error())
	}

	notes := ""
	for k, v := range files {
		if strings.HasSuffix(k, notesFileSuffix) {
			//			fmt.Println(v)
			if k == path.Join(chart.Metadata.Name, "templates", notesFileSuffix) {
				notes = v
			}
			delete(files, k)
		} else {
			fmt.Printf("=========%v=======\n", k)
			fmt.Println(v)
		}
	}
	fmt.Println("notes: ==================")
	fmt.Println(notes)
}
