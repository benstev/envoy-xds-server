//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package v1alpha1

type EnvoyConfig struct {
	Name           string `yaml:"name"`
	Spec           `yaml:"spec"`
	Authenticators []Authenticator
}

type Spec struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type Listener struct {
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
	Routes  []Route `yaml:"routes"`
}

type MatchHeaders struct {
	Name        string `yaml:"name"`
	StringMatch string `yaml:"string_match"`
}

type Match struct {
	Prefix  *string        `yaml:"prefix,omitempty"`
	Path    *string        `yaml:"path,omitempty"`
	Headers []MatchHeaders `yaml:"headers,omitempty"`
}

type Rewrite struct {
	Prefix *string `yaml:"prefix,omitempty"`
}

type Route struct {
	Name         string   `yaml:"name"`
	Match        Match    `yaml:"match"`
	ClusterNames []string `yaml:"clusters"`
	IsGrpc       bool     `yaml:"grpc"`
	Rewrite      *Rewrite `yaml:"rewrite,omitempty"`
	ExternalAuth *bool    `yaml:"external_auth,omitempty"`
}

type Cluster struct {
	Name      string     `yaml:"name"`
	IsGrpc    bool       `yaml:"grpc"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}

type Authenticator struct {
	Name      string   `yaml:"name"`
	Issuer    string   `yaml:"iss"`
	Audiences []string `yaml:"aud"`
	Forward   bool     `yaml:"forward"`
	Secret    string   `yaml:"secret"`
	Match     Match    `yaml:"match"`
}
