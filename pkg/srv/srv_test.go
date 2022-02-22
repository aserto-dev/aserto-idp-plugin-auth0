package srv

import (
	"io"
	"testing"

	"github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/config"
	auth0TestUtils "github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/testutils"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-utils/testutil"
	"github.com/aserto-dev/idp-plugin-sdk/plugin"
	"github.com/stretchr/testify/require"
)

func CreateConfig() config.Auth0Config {
	return config.Auth0Config{
		Domain:       testutil.VaultValue("auth0-idp-test-account.domain"),
		ClientID:     testutil.VaultValue("auth0-idp-test-account.client-id"),
		ClientSecret: testutil.VaultValue("auth0-idp-test-account.client-secret"),
	}
}

func TestOpen(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestWrite(t *testing.T) {
	assert := require.New(t)

	apiUser := auth0TestUtils.CreateTestAPIUser("2ff319e101e1", "Test User", "user@test.com", "https://github.com/aserto-demo/contoso-ad-sample/raw/main/UserImages/Euan%20Garden.jpg")
	cfg := CreateConfig()
	err := cfg.Validate(plugin.OperationTypeWrite)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeWrite)
	assert.Nil(err)

	err = auth0Plugin.Write(apiUser)
	assert.Nil(err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.NotNil(stats)
	assert.Equal(int32(1), stats.Received)
	assert.Equal(int32(1), stats.Created)
	assert.Equal(int32(0), stats.Errors)
}

func TestReadInvalidUserID(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	cfg.UserPID = "somerandomID"
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(0, len(users))

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestReadUserByID(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	cfg.UserPID = "2ff319e101e1"
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.Nil(err)
	assert.Equal(1, len(users))
	assert.Equal(users[0].GetId(), "2ff319e101e1")
	assert.Equal(users[0].GetDisplayName(), "Test User")

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestReadInvalidUserEmail(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	cfg.UserEmail = "invalidID"
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(0, len(users))

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestReadUserByEmail(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	cfg.UserEmail = "user@test.com"
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.Nil(err)
	assert.Equal(1, len(users))
	assert.Equal(users[0].GetEmail(), "user@test.com")
	assert.Equal(users[0].GetDisplayName(), "Test User")

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestRead(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
	assert.Nil(err)

	users, err := auth0Plugin.Read()
	assert.Nil(err)
	assert.Equal(3, len(users))

	_, err = auth0Plugin.Read()
	assert.NotNil(err)
	assert.Equal(io.EOF, err)

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}

func TestDelete(t *testing.T) {
	assert := require.New(t)

	cfg := CreateConfig()
	err := cfg.Validate(plugin.OperationTypeRead)
	assert.Nil(err)

	auth0Plugin := NewAuth0Plugin()
	err = auth0Plugin.Open(&cfg, plugin.OperationTypeRead)
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

	stats, err := auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)

	err = auth0Plugin.Open(&cfg, plugin.OperationTypeDelete)
	assert.Nil(err)

	err = auth0Plugin.Delete("auth0|" + testUser.Id)
	assert.Nil(err)

	stats, err = auth0Plugin.Close()
	assert.Nil(err)
	assert.Nil(stats)
}
