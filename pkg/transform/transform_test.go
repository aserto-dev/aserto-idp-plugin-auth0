package transform

import (
	"reflect"
	"testing"

	auth0TestUtils "github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/testutils"
	"github.com/stretchr/testify/require"
	"gopkg.in/auth0.v5/management"
)

func TestTransformToAuth0(t *testing.T) {
	assert := require.New(t)
	apiUser := auth0TestUtils.CreateTestAPIUser("1", "Name", "email", "pic")

	auth0User := ToAuth0(apiUser, WithUserID())

	assert.True(reflect.TypeOf(*auth0User) == reflect.TypeOf(management.User{}), "the returned object should be *management.User")
	assert.Equal("Name", (*auth0User).GetNickname(), "should correctly detect the nickname")
	assert.Equal("email", (*auth0User).GetEmail(), "should correctly populate the email")
	assert.Equal("pic", (*auth0User).GetPicture(), "should correctly populate the email")
	assert.Equal("1", (*auth0User).GetID(), "should populate id")
}

func TestTransformToAuth0NoID(t *testing.T) {
	assert := require.New(t)
	apiUser := auth0TestUtils.CreateTestAPIUser("1", "Name", "email", "pic")

	auth0User := ToAuth0(apiUser)

	assert.True(reflect.TypeOf(*auth0User) == reflect.TypeOf(management.User{}), "the returned object should be *management.User")
	assert.Equal("Name", (*auth0User).GetNickname(), "should correctly detect the nickname")
	assert.Equal("email", (*auth0User).GetEmail(), "should correctly populate the email")
	assert.Equal("pic", (*auth0User).GetPicture(), "should correctly populate the email")
	assert.Equal("", (*auth0User).GetID(), "should not populate id")
}

func TestTransform(t *testing.T) {
	assert := require.New(t)
	auth0User := auth0TestUtils.CreateTestAuth0User("1", "Name", "email", "pic", "+40722332233", "userName")

	apiUser := Transform(auth0User)

	assert.Empty(apiUser.Id, "should not populate the id")
	assert.Equal("Name", apiUser.DisplayName, "should correctly detect the displayname")
	assert.Equal("email", apiUser.Email, "should correctly populate the email")
	assert.Equal("pic", apiUser.Picture, "should correctly populate the email")
	assert.Equal(4, len(apiUser.Identities))
	assert.Equal("auth0", apiUser.Identities["userName"].Provider)
	assert.False(apiUser.Identities["userName"].Verified)
	assert.Equal("+40722332233", apiUser.Attributes.Properties.Fields["phoneNumber"].GetStringValue())
}
