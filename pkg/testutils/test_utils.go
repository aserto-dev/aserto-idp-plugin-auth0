package testutils_test

import (
	"strings"
	"time"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateTestAPIUser(id, displayName, email, picture string) *api.User {
	user := api.User{
		Id:          id,
		DisplayName: displayName,
		Email:       email,
		Picture:     picture,
		Identities:  make(map[string]*api.IdentitySource),
		Attributes: &api.AttrSet{
			Properties:  &structpb.Struct{Fields: make(map[string]*structpb.Value)},
			Roles:       []string{},
			Permissions: []string{},
		},
		Applications: make(map[string]*api.AttrSet),
		Metadata: &api.Metadata{
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: timestamppb.New(time.Now()),
		},
	}

	return &user
}

func CreateTestAuth0User(id, displayName, email, picture, phoneNo, userName string) *management.User {

	metadata := make(map[string]interface{})
	metadata["phoneNumber"] = phoneNo
	metadata[strings.ToLower(api.IdentityKind_IDENTITY_KIND_PHONE.String())] = phoneNo
	metadata[strings.ToLower(api.IdentityKind_IDENTITY_KIND_USERNAME.String())] = userName

	user := management.User{
		ID:           auth0.String(id),
		Nickname:     auth0.String(displayName),
		Email:        auth0.String(email),
		Picture:      auth0.String(picture),
		UserMetadata: &metadata,
	}

	return &user
}
