package srv

import (
	"strings"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gopkg.in/auth0.v5"
	"gopkg.in/auth0.v5/management"
)

const (
	provider = "auth0"
)

func TransformToAuth0(in *api.User, usernameReq bool, max, min int) (*management.User, error) {
	// TODO: add more data here
	user := management.User{
		ID:           auth0.String(in.Id),
		Nickname:     auth0.String(in.DisplayName),
		Email:        auth0.String(in.Email),
		Picture:      auth0.String(in.Picture),
		UserMetadata: make(map[string]interface{}),
	}

	var username string
	for key, value := range in.Identities {
		if value.Kind == api.IdentityKind_IDENTITY_KIND_USERNAME {
			username = key
		}
	}

	if usernameReq {
		if username != "" {
			runes := []rune(username)

			if len(runes) < max && len(runes) > min {
				user.Username = auth0.String(username)
			} else {
				return nil, status.Errorf(codes.Internal, "username for user %s doesn't have the correct length %d - %d", in.DisplayName, min, max)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "username required is enabled, failed to populate username for user %s", in.DisplayName)
		}

	}

	return &user, nil
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
		Provider: provider,
		Verified: true,
	}

	user.Identities[in.GetEmail()] = &api.IdentitySource{
		Kind:     api.IdentityKind_IDENTITY_KIND_EMAIL,
		Provider: provider,
		Verified: in.GetEmailVerified(),
	}

	phoneProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_PHONE.String())
	if in.UserMetadata[phoneProp] != nil {
		phone := in.UserMetadata[phoneProp].(string)
		user.Identities[phone] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_PHONE,
			Provider: provider,
			Verified: false,
		}
	}

	usernameProp := strings.ToLower(api.IdentityKind_IDENTITY_KIND_USERNAME.String())
	if in.UserMetadata[usernameProp] != nil {
		username := in.UserMetadata[usernameProp].(string)
		user.Identities[username] = &api.IdentitySource{
			Kind:     api.IdentityKind_IDENTITY_KIND_USERNAME,
			Provider: provider,
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
