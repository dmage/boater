package manifests

import (
	"github.com/dmage/boater/pkg/printer"
)

type FSLayer struct {
	BlobSum string `json:"blobSum"`
}

func (fsl FSLayer) Dump(prefix string) {
	printer.KeyValueln(prefix, "blobSum", fsl.BlobSum)
}

type Schema1History struct {
	V1Compatibility string `json:"v1Compatibility"`
}

func (h Schema1History) Dump(prefix string) {
	printer.KeyValueln(prefix, "v1Compatibility", h.V1Compatibility)
}

type Schema1 struct {
	Name         string           `json:"name"`
	Tag          string           `json:"tag"`
	Architecture string           `json:"architecture"`
	FSLayers     []FSLayer        `json:"fsLayers"`
	History      []Schema1History `json:"history"`
}

func (s Schema1) Dump(prefix string) {
	if s.Name != "" {
		printer.KeyValueln(prefix, "name", s.Name)
	}
	if s.Tag != "" {
		printer.KeyValueln(prefix, "tag", s.Tag)
	}
	if s.Architecture != "" {
		printer.KeyValueln(prefix, "architecture", s.Architecture)
	}
	if len(s.FSLayers) > 0 {
		printer.Keyln(prefix, "fsLayers")
		for _, layer := range s.FSLayers {
			layer.Dump(prefix + "- ")
		}
	}
	if len(s.History) > 0 {
		printer.Keyln(prefix, "history")
		for _, history := range s.History {
			history.Dump(prefix + "- ")
		}
	}
}
