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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomnomnom/linkheader"

	"github.com/dmage/boater/pkg/client"
)

type limitedReader struct {
	r io.Reader
	n int64
}

var errResponseTooLarge = fmt.Errorf("response too large")

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if l.n <= 0 {
		return 0, errResponseTooLarge
	}
	if int64(len(p)) > l.n {
		p = p[0:l.n]
	}
	n, err = l.r.Read(p)
	l.n -= int64(n)
	return
}

func getTags(c *client.Client, url string) ([]string, string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected response from registry: %s", resp.Status)
	}

	buf, err := ioutil.ReadAll(&limitedReader{
		r: resp.Body,
		n: 20 << 20, // 20 megabytes
	})
	if err != nil {
		return nil, "", err
	}

	if rootCmdVerbose {
		log.Println(string(buf))
	}

	var response struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	err = json.Unmarshal(buf, &response)
	if err != nil {
		return nil, "", err
	}

	links := linkheader.ParseMultiple(resp.Header.Values("Link"))
	nextLink := ""
	for _, link := range links {
		if link.Rel == "next" {
			nextLink = link.URL
			break
		}
	}

	return response.Tags, nextLink, nil
}

var getTagsCmd = &cobra.Command{
	Use:   "get-tags <repository>",
	Short: "List tags in a repository",
	Long: `List tags in a repository.

Examples:
  # List tags in the repository busybox.
  boater get-tags busybox
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		c := newClient(args[0], []string{"pull"})

		allTags := []string{}
		tagsURL := c.URL("/v2/%s/tags/list", c.Scope())
		for {
			tags, nextURL, err := getTags(c, tagsURL)
			if err != nil {
				log.Fatal(err)
			}

			allTags = append(allTags, tags...)
			if nextURL == "" {
				break
			}

			base, err := url.Parse(tagsURL)
			if err != nil {
				log.Fatal(err)
			}

			ref, err := url.Parse(nextURL)
			if err != nil {
				log.Fatal(err)
			}

			tagsURL = base.ResolveReference(ref).String()
		}
		for _, tag := range allTags {
			fmt.Printf("%s\n", tag)
		}
	},
}

func init() {
	RootCmd.AddCommand(getTagsCmd)
}
