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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/dmage/boater/pkg/client"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token <hostname> [<scope1>,<scope2>,...]",
	Short: "Get a token",
	Long: `Get a bearer token from a registry.

Examples:
  # Get a token from docker.io for docker.io/library/busybox.
  boater token docker.io repository:library/busybox:pull
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			os.Exit(1)
		}

		host := args[0]
		scope := args[1:]

		httpClient := &http.Client{
			Transport: newTransport(),
		}

		schemes := []string{"https"}
		if rootCmdInsecure {
			schemes = append(schemes, "http")
		}

		var resp *http.Response
		var err error
		for _, scheme := range schemes {
			resp, err = httpClient.Get(client.URL(scheme, host, "/v2/"))
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		challenges := challenge.ResponseChallenges(resp)
		if len(challenges) != 1 {
			log.Fatal("unexpected challenges value: ", challenges)
		}

		challenge := challenges[0]
		if challenge.Scheme != "bearer" {
			log.Fatal("unexpected challenge scheme: ", challenge.Scheme)
		}

		realm, ok := challenge.Parameters["realm"]
		if !ok {
			log.Fatal("no realm parameter in the challenge")
		}

		params := url.Values{}
		if service, ok := challenge.Parameters["service"]; ok {
			params["service"] = []string{service}
		}
		if len(scope) > 0 {
			params["scope"] = scope
		}

		_, err = url.Parse(realm)
		if err != nil {
			log.Fatal("parse realm: ", err)
		}

		req, err := http.NewRequest("GET", realm+"?"+params.Encode(), nil)
		if err != nil {
			log.Fatal(err)
		}

		password, havePassword := getPassword()
		if rootCmdUser != "" || havePassword {
			userpass := fmt.Sprintf("%s:%s", rootCmdUser, password)
			token := base64.StdEncoding.EncodeToString([]byte(userpass))
			req.Header.Add("Authorization", "Basic "+token)
		}

		resp, err = httpClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal(resp.Status)
		}

		var v struct {
			Token string
		}
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(v.Token)
	},
}

func init() {
	RootCmd.AddCommand(tokenCmd)
}
