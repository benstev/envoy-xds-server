package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	jwtauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	"github.com/lestrrat-go/jwx/v2/jwk"
	api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"
	"github.com/stevesloka/envoy-xds-server/internal/matcher"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type (
	Config struct {
		Authenticators []Authenticator `yaml:"authenticators"`
	}

	Authenticators struct {
		authenticators []Authenticator
		apikeysUrl     string
	}

	Authenticator struct {
		Name      string   `json:"name"`
		Issuer    string   `json:"iss"`
		Audiences []string `json:"aud"`
		// Name      string   `yaml:"name"`
		// Issuer    string   `yaml:"iss"`
		// Audiences []string `yaml:"aud"`
		// Forward bool `yaml:"forward"`
		// Secret    string    `yaml:"secret"`
		// Match api.Match `yaml:"match"`
	}

	JSONWebKeySet struct {
		Keys []jwk.Key `json:"keys"`
	}
)

const (
	AUTHS_ENDPOINT = "/authenticators"
)

func NewAuthenticators(apikeysUrl string) *Authenticators {
	return &Authenticators{make([]Authenticator, 0), apikeysUrl}
}

func (c *Authenticators) Load() *Authenticators {
	resp, err := http.Get(c.apikeysUrl + AUTHS_ENDPOINT)
	if err != nil || resp.StatusCode != 200 {
		log.Panicf("can't fetch auth records, err : %s", err.Error())
		return nil
	}

	defer resp.Body.Close()

	auths := make([]Authenticator, 0)
	if err := json.NewDecoder(resp.Body).Decode(&auths); err != nil {
		log.Panicf("can't decode auth record, err : %s", err.Error())
		return nil
	}

	c.authenticators = auths
	return c
}

func (c *Authenticators) remoteJwkSource(a Authenticator) *jwtauth.JwtProvider_RemoteJwks {
	return &jwtauth.JwtProvider_RemoteJwks{
		RemoteJwks: &jwtauth.RemoteJwks{
			HttpUri: &core.HttpUri{
				Uri: fmt.Sprintf("%s/jwks/%s", c.apikeysUrl, a.Name),
				HttpUpstreamType: &core.HttpUri_Cluster{
					Cluster: "ext-authz-http",
				},
				Timeout: durationpb.New(60 * time.Second), // BST
			},
			CacheDuration: durationpb.New(1 * time.Second), // BST
		},
	}
}

func (c *Authenticators) providers() map[string]*jwtauth.JwtProvider {
	providers := make(map[string]*jwtauth.JwtProvider)
	for _, a := range c.authenticators {

		providers[a.Name] = &jwtauth.JwtProvider{
			Issuer:    a.Issuer,
			Audiences: a.Audiences,
			// Forward:             true,
			JwksSourceSpecifier: c.remoteJwkSource(a),
		}
	}
	return providers
}

func (c *Authenticators) authMatcher(a Authenticator) api.Match {
	pfx := fmt.Sprintf("/%s/", a.Name)
	return api.Match{
		Prefix: &pfx,
	}
}

func (c *Authenticators) rules() []*jwtauth.RequirementRule {
	rules := make([]*jwtauth.RequirementRule, 0)
	for _, a := range c.authenticators {

		rules = append(rules, &jwtauth.RequirementRule{
			Match: matcher.MakeMatch(c.authMatcher(a)),
			RequirementType: &jwtauth.RequirementRule_Requires{
				Requires: &jwtauth.JwtRequirement{
					RequiresType: &jwtauth.JwtRequirement_ProviderName{
						ProviderName: a.Name,
					},
				},
			},
		})

	}
	return rules
}

func (c *Authenticators) Config() *anypb.Any {
	conf := &jwtauth.JwtAuthentication{
		Providers: c.providers(),
		Rules:     c.rules(),
	}
	authConfig, _ := anypb.New(conf)
	return authConfig
}
