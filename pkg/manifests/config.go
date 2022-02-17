package manifests

import (
	"regexp"
	"time"

	"github.com/dmage/boater/pkg/printer"
)

var reNewLine = regexp.MustCompile(`( {4,}|[ \t]*\t[ \t]*)|\n`)

type HealthConfig struct {
	Test     []string
	Interval int
	Timeout  int
	Retries  int
}

func (hc HealthConfig) Dump(prefix string) {
	if len(hc.Test) > 0 {
		printer.Keyln(prefix, "test")
		for _, test := range hc.Test {
			printer.Valueln(prefix+"- ", test)
		}
	}
	if hc.Interval != 0 {
		printer.KeyValueln(prefix, "interval", hc.Interval)
	}
	if hc.Timeout != 0 {
		printer.KeyValueln(prefix, "timeout", hc.Timeout)
	}
	if hc.Retries != 0 {
		printer.KeyValueln(prefix, "retries", hc.Retries)
	}
}

type ContainerConfig struct {
	User         string
	Memory       int64
	MemorySwap   int64
	CpuShares    int64
	ExposedPorts map[string]struct{}
	Env          []string
	Entrypoint   []string
	Cmd          []string
	Healthcheck  HealthConfig
	Volumes      map[string]struct{}
	WorkingDir   string
	Labels       map[string]string
	StopSignal   string
}

func (cc ContainerConfig) Dump(prefix string) {
	if cc.User != "" {
		printer.KeyValueln(prefix, "User", cc.User)
	}
	if cc.Memory != 0 {
		printer.KeyValueln(prefix, "Memory", cc.Memory)
	}
	if cc.MemorySwap != 0 {
		printer.KeyValueln(prefix, "MemorySwap", cc.MemorySwap)
	}
	if cc.CpuShares != 0 {
		printer.KeyValueln(prefix, "CpuShares", cc.CpuShares)
	}
	if len(cc.ExposedPorts) > 0 {
		printer.Keyln(prefix, "ExposedPorts")
		for port := range cc.ExposedPorts {
			printer.Valueln(prefix+"- ", port)
		}
	}
	if len(cc.Env) > 0 {
		printer.Keyln(prefix, "Env")
		for _, env := range cc.Env {
			printer.Valueln(prefix+"- ", env)
		}
	}
	if len(cc.Entrypoint) > 0 {
		printer.Keyln(prefix, "Entrypoint")
		for _, entrypoint := range cc.Entrypoint {
			printer.Valueln(prefix+"- ", entrypoint)
		}
	}
	if len(cc.Cmd) > 0 {
		printer.Keyln(prefix, "Cmd")
		for _, cmd := range cc.Cmd {
			printer.Valueln(prefix+"- ", cmd)
		}
	}
	if len(cc.Healthcheck.Test) > 0 {
		printer.Keyln(prefix, "Healthcheck")
		cc.Healthcheck.Dump(prefix + "  ")
	}
	if len(cc.Volumes) > 0 {
		printer.Keyln(prefix, "Volumes")
		for volume := range cc.Volumes {
			printer.Valueln(prefix+"- ", volume)
		}
	}
	if cc.WorkingDir != "" {
		printer.KeyValueln(prefix, "WorkingDir", cc.WorkingDir)
	}
	if len(cc.Labels) > 0 {
		printer.Keyln(prefix, "Labels")
		for key, value := range cc.Labels {
			printer.KeyValueln(prefix+"  ", key, value)
		}
	}
	if cc.StopSignal != "" {
		printer.KeyValueln(prefix, "StopSignal", cc.StopSignal)
	}
}

type RootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

func (rfs RootFS) Dump(prefix string) {
	printer.KeyValueln(prefix, "type", rfs.Type)
	printer.Keyln(prefix, "diff_ids")
	for _, diffID := range rfs.DiffIDs {
		printer.Valueln(prefix+"- ", diffID)
	}
}

type History struct {
	Created    time.Time `json:"created"`
	Author     string    `json:"author"`
	CreatedBy  string    `json:"created_by"`
	Comment    string    `json:"comment"`
	EmptyLayer bool      `json:"empty_layer"`
}

func (h History) Dump(prefix string, secondPrefix string) {
	if !h.Created.IsZero() {
		printer.KeyValueln(prefix, "created", h.Created.Format(time.RFC3339))
		prefix = secondPrefix
	}
	if h.Author != "" {
		printer.KeyValueln(prefix, "author", h.Author)
		prefix = secondPrefix
	}
	if h.CreatedBy != "" {
		printer.KeyValueln(prefix, "created_by", reNewLine.ReplaceAllString(h.CreatedBy, "\n"+prefix+"  $1"))
		prefix = secondPrefix
	}
	if h.Comment != "" {
		printer.KeyValueln(prefix, "comment", h.Comment)
		prefix = secondPrefix
	}
	if h.EmptyLayer {
		printer.KeyValueln(prefix, "empty_layer", "true")
		prefix = secondPrefix
	}
}

type ImageConfig struct {
	Created      time.Time       `json:"created"`
	Author       string          `json:"author"`
	Architecture string          `json:"architecture"`
	OS           string          `json:"os"`
	Config       ContainerConfig `json:"config"`
	RootFS       RootFS          `json:"rootfs"`
	History      []History       `json:"history"`
}

func (ic ImageConfig) Dump(prefix string) {
	if !ic.Created.IsZero() {
		printer.KeyValueln(prefix, "created", ic.Created.Format(time.RFC3339))
	}
	if ic.Author != "" {
		printer.KeyValueln(prefix, "author", ic.Author)
	}
	if ic.Architecture != "" {
		printer.KeyValueln(prefix, "architecture", ic.Architecture)
	}
	if ic.OS != "" {
		printer.KeyValueln(prefix, "os", ic.OS)
	}
	printer.Keyln(prefix, "config")
	ic.Config.Dump(prefix + "  ")
	printer.Keyln(prefix, "rootfs")
	ic.RootFS.Dump(prefix + "  ")
	printer.Keyln(prefix, "history")
	for _, history := range ic.History {
		history.Dump(prefix+"- ", prefix+"  ")
	}
}
