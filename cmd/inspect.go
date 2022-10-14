// Copyright © 2022 Oleg Bulatov <oleg@bulatov.me>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dmage/boater/pkg/client"
	"github.com/dmage/boater/pkg/manifests"
	"github.com/dmage/boater/pkg/printer"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <name>[:<tag>|@<digest>]",
	Short: "Inspect a manifest and its config",
	Long: `Inspect an image manifest and its config.

Gets the manifest and its config from the registry and prints them in a
human-readable format to stdout.

NOTE: Output of this command is intended for humans, so it is not guaranteed to
be backward-compatible.

Examples:
  # Inspect a manifest for busybox.
  boater inspect busybox

  # Inspect the manifest by its digest.
  boater inspect busybox@sha256:ee44b399df993016003bf5466bd3eeb221305e9d0fa831606bc7902d149c775b
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		c := newClient(args[0], []string{"pull"})
		repoName := c.Named().Name()
		tagOrDigest := manifestName(c.Named())

		resp, err := c.GetManifest(tagOrDigest, client.GetManifestOptions{
			AcceptKnown: true,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			_, err = io.Copy(os.Stderr, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Fatal(resp.Status)
		}

		manifestType := resp.Header.Get("Content-Type")
		contentDigest := resp.Header.Get("Docker-Content-Digest")
		switch manifestType {
		case "application/vnd.docker.distribution.manifest.v1+json",
			"application/vnd.docker.distribution.manifest.v1+prettyjws":
			var manifest manifests.Schema1
			err = json.NewDecoder(resp.Body).Decode(&manifest)
			if err != nil {
				log.Fatal(err)
			}
			printer.Referencef("%s:%s\n", repoName, manifest.Tag)
			printer.KeyValueln("  ", "Content-Type", manifestType)
			manifest.Dump("  ")
		case "application/vnd.docker.distribution.manifest.v2+json":
			var manifest manifests.Schema2
			err = json.NewDecoder(resp.Body).Decode(&manifest)
			if err != nil {
				log.Fatal(err)
			}
			configResp, err := c.GetBlob(manifest.Config.Digest.String())
			if err != nil {
				log.Fatal(err)
			}
			defer configResp.Body.Close()
			var config manifests.ImageConfig
			err = json.NewDecoder(configResp.Body).Decode(&config)
			if err != nil {
				log.Fatal(err)
			}
			printer.Referencef("%s@%s\n", repoName, contentDigest)
			printer.KeyValueln("  ", "Content-Type", manifestType)
			manifest.Dump("  ", config)
		case "application/vnd.docker.distribution.manifest.list.v2+json":
			var manifest manifests.ManifestList
			err = json.NewDecoder(resp.Body).Decode(&manifest)
			if err != nil {
				log.Fatal(err)
			}
			printer.Referencef("%s@%s\n", repoName, contentDigest)
			printer.KeyValueln("  ", "Content-Type", manifestType)
			manifest.Dump("  ", repoName)
		default:
			log.Fatalf("unsupported manifest type: %s", manifestType)
		}
	},
}

func init() {
	RootCmd.AddCommand(inspectCmd)
}
