package auth

import (
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ExtAuthConfig() *anypb.Any {
	config := &extauth.ExtAuthz{
		Services: &extauth.ExtAuthz_GrpcService{
			GrpcService: &core.GrpcService{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
						ClusterName: "ext-authz",
					},
				},
				Timeout: durationpb.New(3 * time.Second),
			},
		},
		FailureModeAllow:    false,
		TransportApiVersion: core.ApiVersion_V3,
	}

	authConfig, err := anypb.New(config)
	if err != nil {
		panic(err)
	}
	return authConfig
}

func ExtAuthRouteDisabled() map[string]*anypb.Any {
	filterConfig := make(map[string]*anypb.Any)

	extAuthDisabled := &extauth.ExtAuthzPerRoute{
		Override: &extauth.ExtAuthzPerRoute_Disabled{
			Disabled: true,
		},
	}

	if pbst, err := anypb.New(extAuthDisabled); err == nil {
		filterConfig[wellknown.HTTPExternalAuthorization] = pbst
	} else {
		panic(err)
	}
	return filterConfig
}

func ExtAuthRouteSettings(name string) map[string]*anypb.Any {
	filterConfig := make(map[string]*anypb.Any)

	extAuthSettings := &extauth.ExtAuthzPerRoute{
		Override: &extauth.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &extauth.CheckSettings{
				ContextExtensions: map[string]string{"route": name},
			},
		},
	}

	if pbst, err := anypb.New(extAuthSettings); err == nil {
		filterConfig[wellknown.HTTPExternalAuthorization] = pbst
	} else {
		panic(err)
	}
	return filterConfig
}
