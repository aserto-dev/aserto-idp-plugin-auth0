package config

import (
	"github.com/aserto-dev/idp-plugin-sdk/plugin"
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
	Domain         string `description:"Auth0 domain" kind:"attribute" mode:"normal" readonly:"false" name:"domain"`
	ClientID       string `description:"Auth0 Client ID" kind:"attribute" mode:"normal" readonly:"false" name:"client-id"`
	ClientSecret   string `description:"Auth0 Client Secret" kind:"attribute" mode:"normal" readonly:"false" name:"client-secret"`
	ConnectionName string `description:"Auth0 database connection name" kind:"attribute" mode:"normal" readonly:"false" name:"connection-name"`
}

func (c *Auth0Config) Validate(operation plugin.OperationType) error {

	if c.Domain == "" {
		return status.Error(codes.InvalidArgument, "no domain was provided")
	}

	if c.ClientID == "" {
		return status.Error(codes.InvalidArgument, "no client id was provided")
	}

	if c.ClientSecret == "" {
		return status.Error(codes.InvalidArgument, "no client secret was provided")
	}

	if c.ConnectionName == "" {
		c.ConnectionName = "Username-Password-Authentication"
	}

	mgnt, err := management.New(
		c.Domain,
		management.WithClientCredentials(
			c.ClientID,
			c.ClientSecret,
		))

	if err != nil {
		return status.Errorf(codes.Internal, "failed to connect to Auth0, %s", err.Error())
	}

	if operation == plugin.OperationTypeWrite {
		_, err = mgnt.Connection.ReadByName(c.ConnectionName)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get Auth0 connection, %s", err.Error())
		}
	}

	return nil
}

func (c *Auth0Config) Description() string {
	return "Auth0 plugin"
}
