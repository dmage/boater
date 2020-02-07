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
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var putBlobOpts struct {
	Digest string
}

var putBlobCmd = &cobra.Command{
	Use:   "put-blob <imagename> <filename>",
	Short: "Put a blob for an image",
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

		filename := args[1]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		digest := putBlobOpts.Digest
		if digest == "" {
			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}
			digest = fmt.Sprintf("sha256:%x", h.Sum(nil))

			_, err := f.Seek(0, os.SEEK_SET)
			if err != nil {
				log.Fatal(err)
			}
		}

		c := newClient(args[0], []string{"pull", "push"})

		req, err := http.NewRequest("POST", c.URL("/v2/%s/blobs/uploads/", c.Scope()), nil)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			log.Fatalf("%s %s: %s", req.Method, req.URL, resp.Status)
		}

		loc := resp.Header.Get("Location")
		if loc == "" {
			log.Fatal("no Location header")
		}

		uri, err := url.Parse(loc)
		if err != nil {
			log.Fatal("unable to parse Location: %s", err)
		}

		if uri.RawQuery != "" {
			uri.RawQuery += "&"
		}
		uri.RawQuery += "digest=" + url.QueryEscape(digest)

		req, err = http.NewRequest("PUT", uri.String(), f)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/octet-stream")

		resp, err = c.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			log.Fatalf("%s %s: %s", req.Method, req.URL, resp.Status)
		}

		fmt.Println(resp.Header.Get("Docker-Content-Digest"))
	},
}

func init() {
	RootCmd.AddCommand(putBlobCmd)

	putBlobCmd.Flags().StringVar(&putBlobOpts.Digest, "digest", "", "...")
}
