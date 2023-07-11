// Copyright 2020 Envoyproxy Authors
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

import (
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	alog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"

	upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	upstream_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/tcp/v3"

	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	v31 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	v32 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/stevesloka/envoy-xds-server/internal/auth"
	"github.com/stevesloka/envoy-xds-server/internal/matcher"
)

func makeGrpcProtocolOptions() map[string]*anypb.Any {
	protocolOptions := make(map[string]*anypb.Any)

	optionHttp := &upstream_http.HttpProtocolOptions{
		UpstreamProtocolOptions: &upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &core.Http2ProtocolOptions{},
				},
			},
		},
	}

	if pbst, err := anypb.New(optionHttp); err == nil {
		protocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"] = pbst
	} else {
		panic(err)
	}

	optionTcp := &upstream_tcp.TcpProtocolOptions{IdleTimeout: durationpb.New(0 * time.Second)}
	if pbst, err := anypb.New(optionTcp); err == nil {
		protocolOptions["envoy.extensions.upstreams.tcp.v3.TcpProtocolOptions"] = pbst
	} else {
		panic(err)
	}

	return protocolOptions
}

func MakeCluster(clusterName string, grpcCluster bool, eps []Endpoint) *cluster.Cluster {
	cluster := &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       MakeEndpoint(clusterName, eps),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}

	if grpcCluster {
		cluster.TypedExtensionProtocolOptions = makeGrpcProtocolOptions()
	}

	return cluster
}

func MakeEndpoint(clusterName string, eps []Endpoint) *endpoint.ClusterLoadAssignment {
	var endpoints []*endpoint.LbEndpoint

	for _, e := range eps {
		endpoints = append(endpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  e.UpstreamHost,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: e.UpstreamPort,
								},
							},
						},
					},
				},
			},
		})
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: endpoints,
		}},
	}
}

func makeCorsPolicy() map[string]*anypb.Any {
	filterConfig := make(map[string]*anypb.Any)

	filterCorsPolicy := &cors.CorsPolicy{
		AllowOriginStringMatch: []*v32.StringMatcher{{MatchPattern: &v32.StringMatcher_Prefix{Prefix: "*"}}},
		AllowMethods:           "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:           "keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,custom-header-1,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web,grpc-timeout,apikey",
		MaxAge:                 "1728000",
		ExposeHeaders:          "custom-header-1,grpc-status,grpc-message",
	}
	if pbst, err := anypb.New(filterCorsPolicy); err == nil {
		filterConfig[wellknown.CORS] = pbst
	} else {
		panic(err)
	}
	return filterConfig
}

func makeSubRoute(r Route) *route.Route {

	routeAction := &route.RouteAction{
		ClusterSpecifier: &route.RouteAction_Cluster{
			Cluster: r.Cluster,
		},
	}
	if r.IsGrpc {
		routeAction.Timeout = durationpb.New(0 * time.Second)
		routeAction.MaxStreamDuration = &route.RouteAction_MaxStreamDuration{GrpcTimeoutHeaderMax: durationpb.New(0 * time.Second)}
	}

	if r.Rewrite != nil {
		if r.Rewrite.Prefix != nil {
			routeAction.PrefixRewrite = *r.Rewrite.Prefix
		}
	}

	route := &route.Route{
		Match: matcher.MakeMatch(r.Match),
		Action: &route.Route_Route{
			Route: routeAction,
		},
	}

	if !(r.ExternalAuth != nil && *r.ExternalAuth) {
		route.TypedPerFilterConfig = auth.ExtAuthRouteDisabled()
	} else {
		route.TypedPerFilterConfig = auth.ExtAuthRouteSettings(r.Name)
	}
	return route
}

func MakeRoute(routes []Route) *route.RouteConfiguration {
	var rts []*route.Route

	for _, r := range routes {
		rts = append(rts, makeSubRoute(r))
	}

	return &route.RouteConfiguration{
		Name: "listener_0",
		VirtualHosts: []*route.VirtualHost{{
			Name:                 "local_service",
			Domains:              []string{"*"},
			Routes:               rts,
			TypedPerFilterConfig: makeCorsPolicy(),
		}},
	}
}

func makeHTTPConnectionManager(withAccessLog bool, authenticators map[string]auth.Authenticator) *anypb.Any {

	routerConfig, _ := anypb.New(&router.Router{})
	corsConfig, _ := anypb.New(&cors.Cors{})
	grpcWebConfig, _ := anypb.New(&grpcweb.GrpcWeb{})

	manager := &hcm.HttpConnectionManager{
		CodecType:         hcm.HttpConnectionManager_AUTO,
		StatPrefix:        "http",
		StreamIdleTimeout: durationpb.New(0 * time.Second),
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: "listener_0",
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name:       wellknown.CORS,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: corsConfig},
			},
			{
				Name:       wellknown.GRPCWeb,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: grpcWebConfig},
			},
			{
				Name:       wellknown.HTTPExternalAuthorization,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: auth.ExtAuthConfig()},
			},
			{
				Name:       "envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication",
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: auth.JwtAuthConfig(authenticators)},
			},
			{
				Name:       wellknown.Router,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
			},
		},
	}

	if withAccessLog {
		accessLogConfig, _ := anypb.New(&alog.StdoutAccessLog{})
		manager.AccessLog = []*v31.AccessLog{{ConfigType: &v31.AccessLog_TypedConfig{TypedConfig: accessLogConfig}}}
	}

	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return pbst
}

func MakeHTTPListener(listenerName, route, address string, port uint32, withAccessLog bool, authenticators map[string]auth.Authenticator) *listener.Listener {
	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: makeHTTPConnectionManager(withAccessLog, authenticators),
				},
			}},
		}},
	}
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
