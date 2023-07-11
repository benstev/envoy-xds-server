package auth

import (
	"encoding/json"
	"log"
	"os"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	jwtauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	"github.com/stevesloka/envoy-xds-server/internal/matcher"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/lestrrat-go/jwx/v2/jwk"
	api "github.com/stevesloka/envoy-xds-server/apis/v1alpha1"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Authenticators []Authenticator `yaml:"authenticators"`
	}

	Authenticators struct {
		authenticators map[string]Authenticator
	}

	Authenticator struct {
		Name      string    `yaml:"name"`
		Issuer    string    `yaml:"iss"`
		Audiences []string  `yaml:"aud"`
		Forward   bool      `yaml:"forward"`
		Secret    string    `yaml:"secret"`
		Match     api.Match `yaml:"match"`
	}

	JSONWebKeySet struct {
		Keys []jwk.Key `json:"keys"`
	}
)

func NewAuthenticators() *Authenticators { return &Authenticators{make(map[string]Authenticator)} }

func (c *Authenticators) Load(file string) *Authenticators {

	var conf Config

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("error reading JWT Auth YAML file: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatal(err)
	}

	authenticators := make(map[string]Authenticator)
	for _, a := range conf.Authenticators {
		authenticators[a.Name] = a
	}

	c.authenticators = authenticators
	return c
}

func (a *Authenticator) getJwks() []byte {

	raw := []byte(a.Secret)
	key, err := jwk.FromRaw(raw)
	if err != nil {
		os.Exit(1)
	}

	if _, ok := key.(jwk.SymmetricKey); !ok {
		os.Exit(1)
	}

	key.Set(jwk.KeyIDKey, a.Name)

	jwks := JSONWebKeySet{Keys: []jwk.Key{key}}

	b, err := json.Marshal(jwks)
	if err != nil {
		os.Exit(1)
	}
	return b
}

func (c *Authenticators) providers() map[string]*jwtauth.JwtProvider {
	providers := make(map[string]*jwtauth.JwtProvider)
	for k, a := range c.authenticators {

		providers[k] = &jwtauth.JwtProvider{
			Issuer:    a.Issuer,
			Audiences: a.Audiences,
			Forward:   a.Forward,
			JwksSourceSpecifier: &jwtauth.JwtProvider_LocalJwks{
				LocalJwks: &core.DataSource{Specifier: &core.DataSource_InlineBytes{
					InlineBytes: a.getJwks(),
				}},
			},
		}
	}
	return providers
}

func (c *Authenticators) rules() []*jwtauth.RequirementRule {
	rules := make([]*jwtauth.RequirementRule, 0)
	for req, a := range c.authenticators {

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

func (c *Authenticators) Config() *anypb.Any {
	a := &jwtauth.JwtAuthentication{
		Providers: c.providers(),
		Rules:     c.rules(),
	}
	authConfig, _ := anypb.New(a)
	return authConfig
}
