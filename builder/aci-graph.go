package builder

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

func (cnt *Aci) Graph() {
	log.Info("Graph " + cnt.manifest.NameAndVersion)

	os.MkdirAll(cnt.target, 0777)

	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	if cnt.manifest.From != "" {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(cnt.manifest.From.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(cnt.manifest.NameAndVersion.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString("[color=red,penwidth=2.0]")
		buffer.WriteString("\n")
	}

	for _, dep := range cnt.manifest.Aci.Dependencies {
		buffer.WriteString("  ")
		buffer.WriteString("\"")
		buffer.WriteString(dep.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString(" -> ")
		buffer.WriteString("\"")
		buffer.WriteString(cnt.manifest.NameAndVersion.ShortNameId())
		buffer.WriteString("\"")
		buffer.WriteString("\n")
	}

	buffer.WriteString("}\n")

	ioutil.WriteFile(cnt.target+"/graph.dot", buffer.Bytes(), 0644)
}