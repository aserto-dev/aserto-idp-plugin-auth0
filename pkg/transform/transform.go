package transform

import (
	"strings"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	Provider = "auth0"
)

func ToAuth0(in *api.User, setUsername bool) *management.User {
	user := management.User{
		ID:       auth0.String(in.Id),
		Nickname: auth0.String(in.DisplayName),
		Email:    auth0.String(in.Email),
		Picture:  auth0.String(in.Picture),
	}

	if in.Attributes != nil && in.Attributes.Properties != nil {
		user.UserMetadata = in.Attributes.Properties.AsMap()
	}

	if setUsername {
		for key, value := range in.Identities {
			if value.Kind == api.IdentityKind_IDENTITY_KIND_USERNAME {
				user.Username = auth0.String(key)
				break
			}
		}
	}

	return &user
}

// Transform Auth0 user definition into Aserto Edge User object definition.
func Transform(in *management.User) *api.User {

	uid := strings.ToLower(strings.TrimPrefix(*in.ID, "auth0|"))

	user := api.User{
		Id:          uid,
		DisplayName: in.GetNickname(),
		Email:       in.GetEmail(),
		Picture:     in.GetPicture(),
		Identities:  make(map[string]*api.IdentitySource),
		Attributes: &api.AttrSet{
			Properties:  &structpb.Struct{Fields: make(map[string]*structpb.Value)},
			Roles:       []string{},
			Permissions: []string{},
		},
		Applications: make(map[string]*api.AttrSet),
		Metadata: &api.Metadata{
			CreatedAt: timestamppb.New(in.GetCreatedAt()),
			UpdatedAt: timestamppb.New(in.GetUpdatedAt()),
		},
	}

	user.Identities[in.GetID()] = &api.IdentitySource{
		Kind:     api.IdentityKind_IDENTITY_KIND_PID,
		Provider: Provider,
		Verified: true,
	}

	user.Identities[in.GetEmail()] = &api.IdentitySource{
		Kind:     api.IdentityKind_IDENTITY_KIND_EMAIL,
		Provider: Provider,
		Verified: in.GetEmailVerified(),
	}

	phoneProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_PHONE.String())
	if in.UserMetadata[phoneProp] != nil {
		phone := in.UserMetadata[phoneProp].(string)
		user.Identities[phone] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_PHONE,
			Provider: Provider,
			Verified: false,
		}
	}

	usernameProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_USERNAME.String())
	if in.UserMetadata[usernameProp] != nil {
		username := in.UserMetadata[usernameProp].(string)
		user.Identities[username] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_USERNAME,
			Provider: Provider,
			Verified: false,
		}
	}

	if in.UserMetadata != nil && len(in.UserMetadata) != 0 {
		props, err := structpb.NewStruct(in.UserMetadata)
		if err == nil {
			user.Attributes.Properties = props
		}
	}

	return &user
}
