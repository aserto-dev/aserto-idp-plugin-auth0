package transform

import (
	"reflect"
	"testing"

	auth0TestUtils "github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/testutils"

	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/require"
)

func TestTransformToAuth0(t *testing.T) {
	assert := require.New(t)
	apiUser := auth0TestUtils.CreateTestAPIUser("1", "Name", "email", "pic")

	auth0User := ToAuth0(apiUser, false)

	assert.True(
		reflect.TypeOf(*auth0User) == reflect.TypeOf(management.User{}),
		"the returned object should be *management.User",
	)
	assert.Equal("Name", (*auth0User).GetNickname(), "should correctly detect the nickname")
	assert.Equal("email", (*auth0User).GetEmail(), "should correctly populate the email")
	assert.Equal("pic", (*auth0User).GetPicture(), "should correctly populate the email")
}

func TestTransform(t *testing.T) {
	assert := require.New(t)
	auth0User := auth0TestUtils.CreateTestAuth0User(
		"1",
		"Name",
		"email",
		"pic",
		"+40722332233",
		"userName",
	)

	apiUser := Transform(auth0User)

	assert.Equal("1", apiUser.Id, "should correctly populate the id")
	assert.Equal("Name", apiUser.DisplayName, "should correctly detect the displayname")
	assert.Equal("email", apiUser.Email, "should correctly populate the email")
	assert.Equal("pic", apiUser.Picture, "should correctly populate the email")
	assert.Equal(4, len(apiUser.Identities))
	assert.Equal("auth0", apiUser.Identities["userName"].Provider)
	assert.False(apiUser.Identities["userName"].Verified)
	assert.Equal(
		"+40722332233",
		apiUser.Attributes.Properties.Fields["phoneNumber"].GetStringValue(),
	)
}
