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

	"github.com/spf13/cobra"
)

var getBlobCmd = &cobra.Command{
	Use:   "get-blob <imagename> <digest>",
	Short: "Get a blob for an image",
	Long: `Get a blob from an image repository.

Examples:
  # Get the blob from the repository busybox.
  boater get-blob busybox sha256:dc3bacd8b5ea796cea5d6070c8f145df9076f26a6bc1c8981fd5b176d37de843
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}

		c := newClient(args[0], []string{"pull"})
		digest := args[1]

		req, _ := http.NewRequest("GET", c.URL("/v2/%s/blobs/%s", c.Scope(), digest), nil)
		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if _, err = io.Copy(os.Stdout, resp.Body); err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(getBlobCmd)
}
