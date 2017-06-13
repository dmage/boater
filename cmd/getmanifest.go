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

var getManifestOpts struct {
	AcceptSchema1      bool
	AcceptSchema2      bool
	AcceptManifestList bool
	MediaTypes         []string
}

var getManifestCmd = &cobra.Command{
	Use:   "get-manifest <name>[:<tag>|@<digest>]",
	Short: "Get a manifest for an image",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		c := newClient(args[0], []string{"pull"})
		tag := manifestName(c.Named())

		req, err := http.NewRequest("GET", c.URL("/v2/%s/manifests/%s", c.Scope(), tag), nil)
		if err != nil {
			log.Fatal(err)
		}
		if getManifestOpts.AcceptSchema1 {
			req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")
		}
		if getManifestOpts.AcceptSchema2 {
			req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
		}
		if getManifestOpts.AcceptManifestList {
			req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
		}
		for _, mediatype := range getManifestOpts.MediaTypes {
			req.Header.Add("Accept", mediatype)
		}
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
	RootCmd.AddCommand(getManifestCmd)

	getManifestCmd.Flags().BoolVar(&getManifestOpts.AcceptSchema1, "accept-schema1", false, "accept Schema 1 manifests")
	getManifestCmd.Flags().BoolVar(&getManifestOpts.AcceptSchema2, "accept-schema2", false, "accept Schema 2 manifests")
	getManifestCmd.Flags().BoolVar(&getManifestOpts.AcceptManifestList, "accept-manifest-list", false, "accept manifest lists")
	getManifestCmd.Flags().StringArrayVarP(&getManifestOpts.MediaTypes, "mimetype", "t", nil, "accept manifests with a custom MIME type")
}
