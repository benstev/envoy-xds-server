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

package resources

import api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"

type Listener struct {
	Name       string
	Address    string
	Port       uint32
	RouteNames []string
}

type Route struct {
	Name         string
	Match        api.Match
	Cluster      string
	IsGrpc       bool
	Rewrite      *api.Rewrite
	ExternalAuth *bool
}

type Cluster struct {
	Name      string
	IsGrpc    bool
	Endpoints []Endpoint
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

type Authenticator struct {
	Issuer    string
	Audiences []string
	Forward   bool
	Secret    string
	Matches   []api.Match
}
