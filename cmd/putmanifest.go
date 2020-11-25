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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/docker/libtrust"
	"github.com/spf13/cobra"
)

var putManifestOpts struct {
	JSONSignature bool
	MediaType     string
}

var putManifestCmd = &cobra.Command{
	Use:   "put-manifest <imagename>[:<tag>|@<digest>] <filename>",
	Short: "Put a manifest for an image",
	Long: `Put an image manifest into a registry.

Examples:
  # Put the manifest into the repository.
  boater --config-json ~/.docker/config.json put-manifest docker.io/dmage/foo:latest ./manifest.json --content-type="application/vnd.docker.distribution.manifest.v2+json"
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}

		var f io.ReadCloser
		filename := args[1]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if putManifestOpts.JSONSignature {
			pk, err := libtrust.GenerateECP256PrivateKey()
			if err != nil {
				log.Fatal("failed to generate private key for signature: ", err)
			}

			data, err := ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}

			js, err := libtrust.NewJSONSignature(data)
			if err != nil {
				log.Fatal("failed to create json signature: ", err)
			}

			if err := js.Sign(pk); err != nil {
				log.Fatal("failed to sign manifest: ", err)
			}

			pretty, err := js.PrettySignature("signatures")
			if err != nil {
				log.Fatal(err)
			}

			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
			f = ioutil.NopCloser(bytes.NewReader(pretty))
		}

		c := newClient(args[0], []string{"pull", "push"})
		tag := manifestName(c.Named())

		req, err := http.NewRequest("PUT", c.URL("/v2/%s/manifests/%s", c.Scope(), tag), f)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", putManifestOpts.MediaType)

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			io.Copy(os.Stderr, resp.Body)
			os.Stderr.Write([]byte("\n"))
			log.Fatalf("%s %s: %s", req.Method, req.URL, resp.Status)
		}

		fmt.Println(resp.Header.Get("Docker-Content-Digest"))
	},
}

func init() {
	RootCmd.AddCommand(putManifestCmd)

	putManifestCmd.Flags().BoolVarP(&putManifestOpts.JSONSignature, "json-signature", "s", false, "sign the manifest with a random key")
	putManifestCmd.Flags().StringVarP(&putManifestOpts.MediaType, "content-type", "t", "application/vnd.docker.distribution.manifest.v1+json", "use the specified media type to upload the manifest")
}
