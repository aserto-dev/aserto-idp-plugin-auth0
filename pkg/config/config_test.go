package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWithEmptyDomain(t *testing.T) {
	assert := require.New(t)
	config := Auth0Config{
		Domain:       "",
		ClientID:     "id",
		ClientSecret: "secret",
	}

	err := config.Validate()

	assert.NotNil(err)
	assert.Equal("rpc error: code = InvalidArgument desc = no domain was provided", err.Error())
}

func TestValidateWithEmptyClientID(t *testing.T) {
	assert := require.New(t)
	config := Auth0Config{
		Domain:       "domain",
		ClientID:     "",
		ClientSecret: "secret",
	}

	err := config.Validate()

	assert.NotNil(err)
	assert.Equal("rpc error: code = InvalidArgument desc = no client id was provided", err.Error())
}

func TestValidateWithEmptyClientSecret(t *testing.T) {
	assert := require.New(t)
	config := Auth0Config{
		Domain:       "domain",
		ClientID:     "id",
		ClientSecret: "",
	}

	err := config.Validate()

	assert.NotNil(err)
	assert.Equal("rpc error: code = InvalidArgument desc = no client secret was provided", err.Error())
}

// func TestValidateWithInvalidCredentials(t *testing.T) {
// 	assert := require.New(t)
// 	config := Auth0Config{
// 		Domain:       "domain",
// 		ClientID:     "id",
// 		ClientSecret: "secret",
// 	}

// 	err := config.Validate()

// 	assert.NotNil(err)
// 	r, _ := regexp.Compile("Internal desc = failed to get Auth0 connection")
// 	assert.Regexp(r, err.Error())
// }

func TestDescription(t *testing.T) {
	assert := require.New(t)
	config := Auth0Config{
		Domain:       "domain",
		ClientID:     "id",
		ClientSecret: "secret",
	}

	description := config.Description()

	assert.Equal("Auth0 plugin", description)
}
