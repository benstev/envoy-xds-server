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

package xdscache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"

	"github.com/stevesloka/envoy-xds-server/internal/auth"
	"github.com/stevesloka/envoy-xds-server/internal/resources"
)

type XDSCache struct {
	Listeners      map[string]resources.Listener
	Routes         map[string]resources.Route
	Clusters       map[string]resources.Cluster
	Endpoints      map[string]resources.Endpoint
	Authenticators map[string]auth.Authenticator
	WithAccessLog  bool
}

func (xds *XDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeCluster(c.Name, c.IsGrpc, c.Endpoints))
	}

	return r
}

func (xds *XDSCache) RouteContents() []types.Resource {

	var routesArray []resources.Route
	for _, r := range xds.Routes {
		routesArray = append(routesArray, r)
	}

	return []types.Resource{resources.MakeRoute(routesArray)}
}

func (xds *XDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.RouteNames[0], l.Address, l.Port, xds.WithAccessLog, xds.Authenticators))
	}

	return r
}

func (xds *XDSCache) EndpointsContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeEndpoint(c.Name, c.Endpoints))
	}

	return r
}

func (xds *XDSCache) AddAuthenticator(name string, iss string, aud []string, forward bool, secret string, match api.Match) {
	xds.Authenticators[name] = auth.Authenticator{
		Issuer:    iss,
		Audiences: aud,
		Forward:   forward,
		Secret:    secret,
		Match:     match,
	}
}

func (xds *XDSCache) AddListener(name string, routeNames []string, address string, port uint32) {
	xds.Listeners[name] = resources.Listener{
		Name:       name,
		Address:    address,
		Port:       port,
		RouteNames: routeNames,
	}
}

func (xds *XDSCache) AddRoute(name string, match api.Match, clusters []string, isGrpc bool, rewrite *api.Rewrite, externalAuth *bool) {
	xds.Routes[name] = resources.Route{
		Name:         name,
		Match:        match,
		Cluster:      clusters[0],
		IsGrpc:       isGrpc,
		Rewrite:      rewrite,
		ExternalAuth: externalAuth,
	}
}

func (xds *XDSCache) AddCluster(name string, grpc bool) {
	xds.Clusters[name] = resources.Cluster{
		Name:   name,
		IsGrpc: grpc,
	}
}

func (xds *XDSCache) AddEndpoint(clusterName, upstreamHost string, upstreamPort uint32) {
	cluster := xds.Clusters[clusterName]

	cluster.Endpoints = append(cluster.Endpoints, resources.Endpoint{
		UpstreamHost: upstreamHost,
		UpstreamPort: upstreamPort,
	})

	xds.Clusters[clusterName] = cluster
}

func (xds *XDSCache) MakeSnapshot() map[resource.Type][]types.Resource {
	snap := map[resource.Type][]types.Resource{
		// resource.EndpointType: xds.EndpointsContents(),
		resource.ClusterType:  xds.ClusterContents(),
		resource.RouteType:    xds.RouteContents(),
		resource.ListenerType: xds.ListenerContents(),
		// resource.RuntimeType:  []types.Resource{},
		// resource.SecretType:   []types.Resource{},
	}
	return snap
}
