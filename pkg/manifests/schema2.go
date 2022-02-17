package manifests

import (
	"github.com/dmage/boater/pkg/printer"
)

type ConfigDescriptor struct {
	Descriptor
}

func (cd ConfigDescriptor) Dump(prefix string) {
	printer.KeyValueln(prefix, "descriptor", cd.Descriptor)
}

type LayerDescriptor struct {
	Descriptor
	URLs []string `json:"urls"`
}

func (ld LayerDescriptor) Dump(prefix string, secondPrefix string) {
	printer.KeyValueln(prefix, "descriptor", ld.Descriptor)
	prefix = secondPrefix
	if len(ld.URLs) > 0 {
		printer.Keyln(prefix, "urls")
		for _, url := range ld.URLs {
			printer.Valueln(prefix+"- ", url)
		}
	}
}

type Schema2 struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	Config        ConfigDescriptor  `json:"config"`
	Layers        []LayerDescriptor `json:"layers"`
}

func (s Schema2) Dump(prefix string, config ImageConfig) {
	printer.KeyValueln(prefix, "schemaVersion", s.SchemaVersion)
	printer.KeyValueln(prefix, "mediaType", s.MediaType)
	printer.Keyln(prefix, "config")
	s.Config.Dump(prefix + "  ")
	config.Dump(prefix + "  ")
	printer.Keyln(prefix, "layers")
	for _, layer := range s.Layers {
		layer.Dump(prefix+"- ", prefix+"  ")
	}
}
