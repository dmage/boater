// Copyright © 2017 Oleg Bulatov <oleg@bulatov.me>
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
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dmage/boater/pkg/client"
	"github.com/spf13/cobra"
)

var getManifestOpts struct {
	client.GetManifestOptions
}

var getManifestCmd = &cobra.Command{
	Use:   "get-manifest <name>[:<tag>|@<digest>]",
	Short: "Get a manifest for an image",
	Long: `Get an image manifest from a registry.

Prints the manifest to stdout as it came from the registry.

Examples:
  # Get a manifest for busybox.
  boater get-manifest -a busybox

  # Get the manifest by its digest.
  boater get-manifest -a busybox@sha256:ee44b399df993016003bf5466bd3eeb221305e9d0fa831606bc7902d149c775b
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		c := newClient(args[0], []string{"pull"})
		tag := manifestName(c.Named())

		resp, err := c.GetManifest(tag, getManifestOpts.GetManifestOptions)
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

		_, err = io.Copy(os.Stdout, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(getManifestCmd)

	getManifestOpts.GetManifestOptions.AddToFlagSet(getManifestCmd.Flags())
}
