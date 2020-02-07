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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		_, err = io.Copy(os.Stdout, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(getBlobCmd)
}
