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

func ToAuth0(in *api.User, args ...Option) *management.User {
	opts := &transformOptions{}

	for _, arg := range args {
		arg(opts)
	}

	user := management.User{
		Nickname: auth0.String(in.DisplayName),
		Email:    auth0.String(in.Email),
		Picture:  auth0.String(in.Picture),
	}

	if in.Attributes != nil && in.Attributes.Properties != nil {
		properties := in.Attributes.Properties.AsMap()
		user.UserMetadata = &properties
	}

	for key, value := range in.Identities {
		if value.Kind == api.IdentityKind_IDENTITY_KIND_USERNAME {
			user.Username = auth0.String(key)
			break
		}
	}

	if opts.userID {
		user.ID = auth0.String(in.Id)
	}

	return &user
}

// Transform Auth0 user definition into Aserto Edge User object definition.
func Transform(in *management.User) *api.User {
	user := api.User{
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

	if in.UserMetadata == nil {
		return &user
	}

	userMetadata := *in.UserMetadata

	phoneProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_PHONE.String())
	if phone, ok := userMetadata[phoneProp].(string); ok {
		user.Identities[phone] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_PHONE,
			Provider: Provider,
			Verified: false,
		}
	}

	usernameProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_USERNAME.String())
	if username, ok := userMetadata[usernameProp].(string); ok {
		user.Identities[username] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_USERNAME,
			Provider: Provider,
			Verified: false,
		}
	}

	if in.UserMetadata != nil && len(userMetadata) != 0 {
		props, err := structpb.NewStruct(userMetadata)
		if err == nil {
			user.Attributes.Properties = props
		}
	}

	return &user
}
