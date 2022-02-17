package manifests

import (
	"github.com/dmage/boater/pkg/printer"
)

type PlatformSpec struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version"`
	OSFeatures   []string `json:"os.features"`
	Variant      string   `json:"variant"`
	Features     []string `json:"features"`
}

func (ps PlatformSpec) Dump(prefix string) {
	if ps.Architecture != "" {
		printer.KeyValueln(prefix, "architecture", ps.Architecture)
	}
	if ps.OS != "" {
		printer.KeyValueln(prefix, "os", ps.OS)
	}
	if ps.OSVersion != "" {
		printer.KeyValueln(prefix, "os.version", ps.OSVersion)
	}
	if len(ps.OSFeatures) > 0 {
		printer.Keyln(prefix, "os.features")
		for _, feature := range ps.OSFeatures {
			printer.Valueln(prefix+"- ", feature)
		}
	}
	if ps.Variant != "" {
		printer.KeyValueln(prefix, "variant", ps.Variant)
	}
	if len(ps.Features) > 0 {
		printer.Keyln(prefix, "features")
		for _, feature := range ps.Features {
			printer.Valueln(prefix+"- ", feature)
		}
	}
}

type ManifestDescriptor struct {
	Descriptor
	Platform PlatformSpec `json:"platform"`
}

func (md ManifestDescriptor) Dump(prefix string) {
	printer.KeyValueln(prefix, "descriptor", md.Descriptor)
	printer.Keyln(prefix, "platform")
	md.Platform.Dump(prefix + "  ")
}

type ManifestList struct {
	SchemaVersion int                  `json:"schemaVersion"`
	MediaType     string               `json:"mediaType"`
	Manifests     []ManifestDescriptor `json:"manifests"`
}

func (ml ManifestList) Dump(prefix string, repoName string) {
	printer.KeyValueln(prefix, "schemaVersion", ml.SchemaVersion)
	printer.KeyValueln(prefix, "mediaType", ml.MediaType)
	printer.Keyln(prefix, "manifests")
	for _, md := range ml.Manifests {
		printer.Delim(prefix + "  ")
		printer.Referencef("%s@%s\n", repoName, md.Digest)
		md.Dump(prefix + "    ")
	}
}
