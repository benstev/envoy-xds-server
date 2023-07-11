package matcher

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	v32 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"
)

func MakeMatch(m api.Match) *route.RouteMatch {
	match := &route.RouteMatch{}

	if m.Prefix != nil {
		match.PathSpecifier = &route.RouteMatch_Prefix{Prefix: *m.Prefix}
	}

	if m.Path != nil {
		match.PathSpecifier = &route.RouteMatch_Path{Path: *m.Path}
	}

	if len(m.Headers) > 0 {
		headers := make([]*route.HeaderMatcher, 0)
		for _, mhd := range m.Headers {
			headers = append(headers, &route.HeaderMatcher{
				Name: mhd.Name,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &v32.StringMatcher{MatchPattern: &v32.StringMatcher_Exact{Exact: mhd.StringMatch}},
				},
			})
		}
		match.Headers = headers
	}

	return match
}
