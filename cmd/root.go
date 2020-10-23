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
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/credentialprovider"

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
		os.Exit(-1)
	}
}

var rootCmdUser string
var rootCmdPassword string
var rootCmdPasswordFile string
var rootCmdConfigJson string
var rootCmdInsecure bool
var rootCmdVerbose bool

func init() {
	RootCmd.PersistentFlags().StringVarP(&rootCmdUser, "user", "u", "", "use the specified username")
	RootCmd.PersistentFlags().StringVarP(&rootCmdPassword, "password", "p", "", "use the specified password")
	RootCmd.PersistentFlags().StringVarP(&rootCmdPasswordFile, "password-file", "", "", "use the password found in the specified file")
	RootCmd.PersistentFlags().StringVarP(&rootCmdConfigJson, "config-json", "", "", "use credentials from the specified Docker config.json file")
	RootCmd.PersistentFlags().BoolVar(&rootCmdInsecure, "insecure", false, "send requests using http")
	RootCmd.PersistentFlags().BoolVarP(&rootCmdVerbose, "verbose", "v", false, "print http requests")
}

func manifestName(named reference.Named) string {
	if digested, ok := named.(reference.Digested); ok {
		return string(digested.Digest())
	}
	return reference.TagNameOnly(named).(reference.Tagged).Tag()
}

func getPassword() (string, bool) {
	if rootCmdPassword != "" {
		return rootCmdPassword, true
	}

	if rootCmdPasswordFile != "" {
		password, err := ioutil.ReadFile(rootCmdPasswordFile)
		if err != nil {
			log.Fatal(err)
		}

		return strings.TrimRight(string(password), "\r\n"), true
	}

	return "", false
}

func getCredentialsFromConfigJson(configJsonFile string, ref string) ([]credentialprovider.AuthConfig, error) {
	f, err := os.Open(configJsonFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var dockerConfigJSON credentialprovider.DockerConfigJSON
	err = json.NewDecoder(f).Decode(&dockerConfigJSON)
	if err != nil {
		return nil, err
	}

	basicKeyring := &credentialprovider.BasicDockerKeyring{}
	basicKeyring.Add(dockerConfigJSON.Auths)
	creds, _ := basicKeyring.Lookup(ref)
	return creds, nil
}

func newCredentialStore(ref string) auth.CredentialStore {
	password, havePassword := getPassword()
	if rootCmdUser != "" || havePassword {
		if rootCmdVerbose {
			log.Printf("Using credentials provided by command line arguments: %s:%s", rootCmdUser, password)
		}
		return &client.BasicCredentials{
			Username: rootCmdUser,
			Password: password,
		}
	}

	if rootCmdConfigJson != "" {
		if rootCmdVerbose {
			log.Printf("Loading credentials from %s...", rootCmdConfigJson)
		}
		creds, err := getCredentialsFromConfigJson(rootCmdConfigJson, ref)
		if err != nil {
			log.Fatalf("unable to load credentials from %s: %s", rootCmdConfigJson, err)
		}
		if len(creds) > 0 {
			if rootCmdVerbose {
				log.Printf("Using credentials from config.json: %s:%s", creds[0].Username, creds[0].Password)
			}
			return &client.BasicCredentials{
				Username: creds[0].Username,
				Password: creds[0].Password,
			}
		}
	}

	if rootCmdVerbose {
		log.Println("No credentials are found, proceeding as anonymous...")
	}

	return nil
}

func ProxyFromEnvironment(req *http.Request) (*url.URL, error) {
	u, err := http.ProxyFromEnvironment(req)
	if u == nil || err != nil {
		return u, err
	}
	if rootCmdVerbose {
		log.Printf("Using proxy for %s: %s", req.URL, u)
	}
	return u, nil
}

func newTransport() http.RoundTripper {
	t := &http.Transport{
		Proxy: ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if rootCmdInsecure {
		t.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	rt := http.RoundTripper(t)
	if rootCmdVerbose {
		rt = &httplog.RoundTripper{
			RoundTripper: rt,
			N:            256,
		}
	}
	return rt
}

func newClient(ref string, actions []string) *client.Client {
	creds := newCredentialStore(ref)
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
