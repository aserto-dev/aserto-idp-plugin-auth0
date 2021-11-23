package srv

import (
	"io"
	"testing"

	"github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/config"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-utils/testutil"
	"github.com/aserto-dev/idp-plugin-sdk/plugin"
	"github.com/stretchr/testify/require"
	"gopkg.in/auth0.v5/management"
)

func CreateConfig() config.Auth0Config {
	return config.Auth0Config{
		Domain:       testutil.VaultValue("auth0-idp-test-account.domain"),
		ClientID:     testutil.VaultValue("auth0-idp-test-account.client-id"),
		ClientSecret: testutil.VaultValue("auth0-idp-test-account.client-secret"),
	}
}

func CreateAuth0Connection(options interface{}) *management.Connection {
	return &management.Connection{Options: options}
}

func TestOpen(t *testing.T) {
	assert := require.New(t)

	config := CreateConfig()
	err := config.Validate()
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&config, plugin.OperationTypeRead)
	assert.Nil(err)

	err = auth0Plugin.Close()
	assert.Nil(err)
}

func TestWrite(t *testing.T) {
	assert := require.New(t)

	apiUser := CreateTestApiUser("2ff319e101e1", "Test User", "user@test.com", "https://github.com/aserto-demo/contoso-ad-sample/raw/main/UserImages/Euan%20Garden.jpg", "")
	config := CreateConfig()
	err := config.Validate()
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&config, plugin.OperationTypeWrite)
	assert.Nil(err)

	err = auth0Plugin.Write(apiUser)
	assert.Nil(err)

	err = auth0Plugin.Close()
	assert.Nil(err)
}

func TestRead(t *testing.T) {
	assert := require.New(t)

	config := CreateConfig()
	err := config.Validate()
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&config, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.Nil(err)
	assert.Equal(3, len(users))

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	err = auth0Plugin.Close()
	assert.Nil(err)
}

func TestDelete(t *testing.T) {
	assert := require.New(t)

	config := CreateConfig()
	err := config.Validate()
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&config, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.Nil(err)
	assert.Less(1, len(users))

	var testUser *api.User

	for _, user := range users {
		if user.DisplayName == "Test User" {
			testUser = user
		}
	}
	assert.NotNil(testUser)

	err = auth0Plugin.Close()
	assert.Nil(err)

	err = auth0Plugin.Open(&config, plugin.OperationTypeDelete)
	assert.Nil(err)

	err = auth0Plugin.Delete("auth0|" + testUser.Id)
	assert.Nil(err)

	err = auth0Plugin.Close()
	assert.Nil(err)
}

func TestRequiresUsernameWithNilOptions(t *testing.T) {
	assert := require.New(t)
	conn := CreateAuth0Connection(nil)

	required, max, min := requiresUsername(conn)
	assert.False(required)
	assert.Equal(0, max)
	assert.Equal(0, min)
}

func TestRequiresUsernameWithEmptyOptions(t *testing.T) {
	assert := require.New(t)
	connOptions := &management.ConnectionOptions{}
	conn := CreateAuth0Connection(connOptions)

	required, max, min := requiresUsername(conn)
	assert.False(required)
	assert.Equal(0, max)
	assert.Equal(0, min)
}

func TestRequiresUsernameWithOtherValidationOptions(t *testing.T) {
	assert := require.New(t)
	connOptions := &management.ConnectionOptions{Validation: map[string]interface{}{"verify": true}}
	conn := CreateAuth0Connection(connOptions)

	required, max, min := requiresUsername(conn)
	assert.False(required)
	assert.Equal(0, max)
	assert.Equal(0, min)
}

func TestRequiresUsernameWithNoLimitOptions(t *testing.T) {
	assert := require.New(t)
	connOptions := &management.ConnectionOptions{Validation: map[string]interface{}{"username": true}}
	conn := CreateAuth0Connection(connOptions)

	required, max, min := requiresUsername(conn)
	assert.True(required)
	assert.Equal(15, max)
	assert.Equal(1, min)
}

func TestRequiresUsernameWithInvalidLimitOptions(t *testing.T) {
	assert := require.New(t)
	limitMap := map[string]interface{}{"min": "4"}
	connOptions := &management.ConnectionOptions{
		Validation: map[string]interface{}{"username": limitMap},
	}
	conn := CreateAuth0Connection(connOptions)

	required, max, min := requiresUsername(conn)
	assert.True(required)
	assert.Equal(15, max)
	assert.Equal(1, min)
}

func TestRequiresUsernameWithLimitOptions(t *testing.T) {
	assert := require.New(t)
	limitMap := map[string]interface{}{"min": float64(4), "max": float64(20)}
	connOptions := &management.ConnectionOptions{
		Validation: map[string]interface{}{"username": limitMap},
	}
	conn := CreateAuth0Connection(connOptions)

	required, max, min := requiresUsername(conn)
	assert.True(required)
	assert.Equal(20, max)
	assert.Equal(4, min)
}
