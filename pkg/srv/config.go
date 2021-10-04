package srv

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/auth0.v5/management"
)

// values set by linker using ldflag -X
var (
	ver    string // nolint:gochecknoglobals // set by linker
	date   string // nolint:gochecknoglobals // set by linker
	commit string // nolint:gochecknoglobals // set by linker
)

func GetVersion() (string, string, string) {
	return ver, date, commit
}

type Auth0Config struct {
	Domain       string `description:"Auth0 domain" kind:"attribute" mode:"normal" readonly:"false"`
	ClientID     string `description:"Auth0 Client ID" kind:"attribute" mode:"normal" readonly:"false"`
	ClientSecret string `description:"Auth0 Client Secret" kind:"attribute" mode:"normal" readonly:"false"`
}

func (c *Auth0Config) Validate() error {
	mgnt, err := management.New(
		c.Domain,
		management.WithClientCredentials(
			c.ClientID,
			c.ClientSecret,
		))
	if err != nil {
		return status.Error(codes.Internal, "failed to connect to Auth0")
	}

	_, err = mgnt.Connection.ReadByName("Username-Password-Authentication")
	if err != nil {
		return status.Error(codes.Internal, "failed to get Auth0 connection")
	}

	return nil
}

func (c *Auth0Config) Description() string {
	return "Auth0 plugin"
}
