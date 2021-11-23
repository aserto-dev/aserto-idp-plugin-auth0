package srv

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/auth0.v5/management"
)

func TestTransformToAuth0UserNameNotRequired(t *testing.T) {
	assert := require.New(t)
	apiUser := CreateTestApiUser("1", "Name", "email", "pic", "")

	auth0User, err := TransformToAuth0(apiUser, false, 0, 0)

	assert.Nil(err)
	assert.True(reflect.TypeOf(*auth0User) == reflect.TypeOf(management.User{}), "the returned object should be *management.User")
	assert.Equal("Name", (*auth0User).GetNickname(), "should correctly detect the nickname")
	assert.Equal("email", (*auth0User).GetEmail(), "should correctly populate the email")
	assert.Equal("pic", (*auth0User).GetPicture(), "should correctly populate the email")
}

func TestTransformToAuth0UserNameRequired(t *testing.T) {
	assert := require.New(t)
	apiUser := CreateTestApiUser("1", "Name", "email", "pic", "test.user")

	auth0User, err := TransformToAuth0(apiUser, true, 15, 1)

	assert.Nil(err)
	assert.True(reflect.TypeOf(*auth0User) == reflect.TypeOf(management.User{}), "the returned object should be *management.User")
	assert.Equal("Name", (*auth0User).GetNickname(), "should correctly detect the nickname")
	assert.Equal("email", (*auth0User).GetEmail(), "should correctly populate the email")
	assert.Equal("pic", (*auth0User).GetPicture(), "should correctly populate the email")
	assert.Equal("test.user", (*auth0User).GetUsername(), "should populate the username")
}

func TestTransformToAuth0UserNameRequiredWithInvalidLength(t *testing.T) {
	assert := require.New(t)
	apiUser := CreateTestApiUser("1", "Captain", "email", "pic", "test.user")

	auth0User, err := TransformToAuth0(apiUser, true, 25, 10)

	assert.NotNil(err)
	assert.Nil(auth0User)
	assert.Equal("rpc error: code = Internal desc = username for user Captain doesn't have the correct length 10 - 25", err.Error())
}

func TestTransformToAuth0FailToPopulateUserName(t *testing.T) {
	assert := require.New(t)
	apiUser := CreateTestApiUser("1", "Captain America", "email", "pic", "")

	auth0User, err := TransformToAuth0(apiUser, true, 25, 8)

	assert.NotNil(err)
	assert.Nil(auth0User)
	assert.Equal("rpc error: code = Internal desc = username required is enabled, failed to populate username for user Captain America", err.Error())
}

func TestTransform(t *testing.T) {
	assert := require.New(t)
	auth0User := CreateTestAuth0User("1", "Name", "email", "pic", "+40722332233", "userName")

	apiUser := Transform(auth0User)

	assert.Equal("1", apiUser.Id, "should correctly populate the id")
	assert.Equal("Name", apiUser.DisplayName, "should correctly detect the displayname")
	assert.Equal("email", apiUser.Email, "should correctly populate the email")
	assert.Equal("pic", apiUser.Picture, "should correctly populate the email")
	assert.Equal(4, len(apiUser.Identities))
	assert.Equal("auth0", apiUser.Identities["userName"].Provider)
	assert.False(apiUser.Identities["userName"].Verified)
	assert.Equal("+40722332233", apiUser.Attributes.Properties.Fields["phoneNumber"].GetStringValue())
}
