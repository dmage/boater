package manifests

import (
	"fmt"
	"strings"

	"github.com/opencontainers/go-digest"
)

func humanSize(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(b)/1024/1024)
	}
	return fmt.Sprintf("%.2f GB", float64(b)/1024/1024/1024)
}

type Descriptor struct {
	MediaType string        `json:"mediaType"`
	Size      int64         `json:"size"`
	Digest    digest.Digest `json:"digest"`
}

func (d Descriptor) String() string {
	var extra []string
	if d.MediaType != "" {
		extra = append(extra, d.MediaType)
	}
	if d.Size > 0 {
		extra = append(extra, humanSize(d.Size))
	}
	if len(extra) > 0 {
		return fmt.Sprintf("%s (%s)", d.Digest, strings.Join(extra, ", "))
	}
	return d.Digest.String()
}
