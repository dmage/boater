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
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/spf13/cobra"

	"github.com/dmage/boater/pkg/client"
	"github.com/dmage/boater/pkg/httplog"
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "boater",
	Short: "A Docker Registry HTTP API client",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var rootCmdUser string
var rootCmdPassword string
var rootCmdInsecure bool
var rootCmdVerbose bool

func init() {
	RootCmd.PersistentFlags().StringVarP(&rootCmdUser, "user", "u", "", "use the specified username")
	RootCmd.PersistentFlags().StringVarP(&rootCmdPassword, "password", "p", "", "use the specified password")
	RootCmd.PersistentFlags().BoolVar(&rootCmdInsecure, "insecure", false, "send requests using http")
	RootCmd.PersistentFlags().BoolVarP(&rootCmdVerbose, "verbose", "v", false, "print http requests")
}

func manifestName(named reference.Named) string {
	if digested, ok := named.(reference.Digested); ok {
		return string(digested.Digest())
	}
	return reference.TagNameOnly(named).(reference.Tagged).Tag()
}

func newCredentialStore() auth.CredentialStore {
	if rootCmdUser != "" || rootCmdPassword != "" {
		return &client.BasicCredentials{
			Username: rootCmdUser,
			Password: rootCmdPassword,
		}
	}
	return nil
}

func newTransport() http.RoundTripper {
	if rootCmdVerbose {
		return &httplog.RoundTripper{}
	}
	return nil
}

func newClient(ref string, actions []string) *client.Client {
	creds := newCredentialStore()
	transport := newTransport()

	client, err := client.New(ref, rootCmdInsecure, transport)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Auth(creds, client.Scope(), actions...)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
