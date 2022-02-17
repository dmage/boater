package manifests

import (
	"github.com/dmage/boater/pkg/printer"
)

type LayerDescriptor struct {
	Descriptor
	URLs []string `json:"urls"`
}

func (ld LayerDescriptor) Dump(prefix string) {
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
	Config        Descriptor        `json:"config"`
	Layers        []LayerDescriptor `json:"layers"`
}

func (s Schema2) Dump(prefix string, config ImageConfig) {
	printer.KeyValueln(prefix, "schemaVersion", s.SchemaVersion)
	printer.KeyValueln(prefix, "mediaType", s.MediaType)
	printer.KeyValueln(prefix, "config", s.Config.String())
	config.Dump(prefix + "  ")
	printer.Keyln(prefix, "layers")
	for _, layer := range s.Layers {
		printer.Valueln(prefix+"- ", layer.String())
		layer.Dump(prefix + "    ")
	}
}
