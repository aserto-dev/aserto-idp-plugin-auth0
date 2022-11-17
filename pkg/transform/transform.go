package transform

import (
	"strings"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
)

const (
	Provider = "auth0"
)

var (
	phoneProp    = strings.ToLower(api.IdentityKind_IDENTITY_KIND_PHONE.String())
	usernameProp = strings.ToLower(api.IdentityKind_IDENTITY_KIND_USERNAME.String())
)

func ToAuth0(in *api.User) *management.User {
	user := management.User{
		ID:       auth0.String(in.Id),
		Nickname: auth0.String(in.DisplayName),
		Email:    auth0.String(in.Email),
		Picture:  auth0.String(in.Picture),
	}

	if in.Attributes != nil && in.Attributes.Properties != nil {
		props := in.Attributes.Properties.AsMap()
		user.UserMetadata = &props
	}

	for key, value := range in.Identities {
		if value.Kind == api.IdentityKind_IDENTITY_KIND_USERNAME {
			user.Username = auth0.String(key)
			break
		}
	}

	return &user
}

// Transform Auth0 management user into an Aserto v1 User instance.
func Transform(in *management.User) *api.User {

	uid := strings.ToLower(strings.TrimPrefix(*in.ID, "auth0|"))
	if !isValidID(uid) {
		uid = ""
	}

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

	if in.Username != nil {
		user.Identities[in.GetUsername()] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_USERNAME,
			Provider: Provider,
			Verified: true,
		}
	}

	if in.UserMetadata != nil {

		if (*in.UserMetadata)[phoneProp] != nil {
			phone := (*in.UserMetadata)[phoneProp].(string)
			user.Identities[phone] = &api.IdentitySource{
				Kind:     api.IdentityKind_IDENTITY_KIND_PHONE,
				Provider: Provider,
				Verified: false,
			}
		}

		if (*in.UserMetadata)[usernameProp] != nil && in.Username == nil {
			username := (*in.UserMetadata)[usernameProp].(string)
			user.Identities[username] = &api.IdentitySource{
				Kind:     api.IdentityKind_IDENTITY_KIND_USERNAME,
				Provider: Provider,
				Verified: false,
			}
		}

		if len(*in.UserMetadata) != 0 {
			props, err := structpb.NewStruct(*in.UserMetadata)
			if err == nil {
				user.Attributes.Properties = props
			}
		}

	}

	return &user
}

func isValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
