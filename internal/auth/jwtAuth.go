package auth

import (
	"encoding/json"
	"os"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	jwtauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	"github.com/lestrrat-go/jwx/v2/jwk"
	api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"
	"github.com/stevesloka/envoy-xds-server/internal/matcher"
	"google.golang.org/protobuf/types/known/anypb"
)

type (
	Authenticator struct {
		Issuer    string
		Audiences []string
		Forward   bool
		Secret    string
		Match     api.Match
	}

	JSONWebKeySet struct {
		Keys []jwk.Key `json:"keys"`
	}
)

func getJwks(secret string) []byte {

	raw := []byte(secret)
	key, err := jwk.FromRaw(raw)
	if err != nil {
		os.Exit(1)
	}

	if _, ok := key.(jwk.SymmetricKey); !ok {
		os.Exit(1)
	}

	key.Set(jwk.KeyIDKey, "opener-key")

	jwks := JSONWebKeySet{Keys: []jwk.Key{key}}

	b, err := json.Marshal(jwks)
	if err != nil {
		os.Exit(1)
	}
	return b
}

func providers(authenticators map[string]Authenticator) map[string]*jwtauth.JwtProvider {
	providers := make(map[string]*jwtauth.JwtProvider)
	for k, a := range authenticators {

		providers[k] = &jwtauth.JwtProvider{
			Issuer:    a.Issuer,
			Audiences: a.Audiences,
			Forward:   a.Forward,
			JwksSourceSpecifier: &jwtauth.JwtProvider_LocalJwks{
				LocalJwks: &core.DataSource{Specifier: &core.DataSource_InlineBytes{
					InlineBytes: getJwks(a.Secret),
				}},
			},
		}
	}
	return providers
}

func rules(authenticators map[string]Authenticator) []*jwtauth.RequirementRule {
	rules := make([]*jwtauth.RequirementRule, 0)
	for req, a := range authenticators {

		rules = append(rules, &jwtauth.RequirementRule{
			Match: matcher.MakeMatch(a.Match),
			RequirementType: &jwtauth.RequirementRule_Requires{
				Requires: &jwtauth.JwtRequirement{
					RequiresType: &jwtauth.JwtRequirement_ProviderName{
						ProviderName: req,
					},
				},
			},
		})

	}
	return rules
}

func JwtAuthConfig(authenticators map[string]Authenticator) *anypb.Any {
	a := &jwtauth.JwtAuthentication{
		Providers: providers(authenticators),
		Rules:     rules(authenticators),
	}
	authConfig, _ := anypb.New(a)
	return authConfig
}
