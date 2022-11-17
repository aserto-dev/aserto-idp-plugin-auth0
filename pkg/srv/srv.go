package srv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/config"
	"github.com/aserto-dev/aserto-idp-plugin-auth0/pkg/transform"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/idp-plugin-sdk/plugin"
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxBatchSize = int64(500 * 1024)
)

type Auth0Plugin struct {
	Config       *config.Auth0Config
	mgmt         *management.Management
	page         int
	finishedRead bool
	totalSize    int64
	jobs         []management.Job
	users        []map[string]interface{}
	connectionID string
	wg           sync.WaitGroup
	op           plugin.OperationType
}

func NewAuth0Plugin() *Auth0Plugin {
	return &Auth0Plugin{
		Config: &config.Auth0Config{},
	}
}

func (s *Auth0Plugin) GetConfig() plugin.Config {
	return &config.Auth0Config{}
}

func (s *Auth0Plugin) GetVersion() (string, string, string) {
	return config.GetVersion()
}

func (s *Auth0Plugin) Open(cfg plugin.Config, operation plugin.OperationType) error {
	auth0Config, ok := cfg.(*config.Auth0Config)
	if !ok {
		return errors.New("invalid config")
	}

	if auth0Config.UserPID != "" && !strings.HasPrefix(auth0Config.UserPID, "auth0|") {
		auth0Config.UserPID = "auth0|" + auth0Config.UserPID
	}

	s.Config = auth0Config
	s.page = 0
	s.finishedRead = false
	s.op = operation

	mgmt, err := management.New(
		auth0Config.Domain,
		management.WithClientCredentials(
			auth0Config.ClientID,
			auth0Config.ClientSecret,
		))

	if err != nil {
		return nil
	}

	s.mgmt = mgmt

	if operation == plugin.OperationTypeWrite {
		if auth0Config.ConnectionName == "" {
			auth0Config.ConnectionName = "Username-Password-Authentication"
		}

		c, err := mgmt.Connection.ReadByName(auth0Config.ConnectionName)
		if err != nil {
			return err
		}
		s.connectionID = auth0.StringValue(c.ID)
	}

	return nil
}

func (s *Auth0Plugin) Read() ([]*api.User, error) {
	if s.finishedRead {
		return nil, io.EOF
	}

	var errs error
	var users []*api.User

	if s.Config.UserPID != "" {
		user, err := s.readByPID(s.Config.UserPID)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
		return users, nil
	}

	if s.Config.UserEmail != "" {
		return s.readByEmail(s.Config.UserEmail)
	}

	opts := management.Page(s.page)
	ul, err := s.mgmt.User.List(opts)
	if err != nil {
		return nil, err
	}

	for _, u := range ul.Users {
		user := transform.Transform(u)

		users = append(users, user)
	}
	if !ul.HasNext() {
		s.finishedRead = true
	}
	s.page++

	return users, errs
}

func (s *Auth0Plugin) readByPID(id string) (*api.User, error) {

	user, err := s.mgmt.User.Read(id)
	s.finishedRead = true
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("failed to get user by pid %s", id)
	}

	return transform.Transform(user), nil
}

func (s *Auth0Plugin) readByEmail(email string) ([]*api.User, error) {
	var users []*api.User

	auth0Users, err := s.mgmt.User.ListByEmail(email)
	s.finishedRead = true
	if err != nil {
		return nil, err
	}
	if len(auth0Users) < 1 {
		return nil, fmt.Errorf("failed to get user by email %s", email)
	}

	for _, user := range auth0Users {
		apiUser := transform.Transform(user)
		users = append(users, apiUser)
	}

	return users, nil
}

func (s *Auth0Plugin) Write(user *api.User) error {
	u := transform.ToAuth0(user)

	userMap, size, err := structToMap(u)
	if err != nil {
		return err
	}

	if s.totalSize+size < maxBatchSize {
		s.users = append(s.users, userMap)
	} else {
		err = s.startJob()
		if err != nil {
			return err
		}
		s.users = make([]map[string]interface{}, 0)
		s.users = append(s.users, userMap)
	}

	return nil
}

func (s *Auth0Plugin) Delete(userID string) error {
	if s.mgmt == nil {
		return status.Error(codes.Internal, "auth0 management client not initialized")
	}

	return s.mgmt.User.Delete(userID)
}

func (s *Auth0Plugin) Close() (*plugin.Stats, error) {
	switch s.op { //nolint : gocritic // tbd
	case plugin.OperationTypeWrite:
		if len(s.users) > 0 {
			err := s.startJob()

			if err != nil {
				return nil, err
			}
		}

		var errs error
		stats := &plugin.Stats{}
		for i := 0; i < len(s.jobs); i++ {
			jobID := auth0.StringValue(s.jobs[i].ID)
			err := s.waitJob(jobID)
			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				auth0Stats, err := retrieveJobSummary(s.mgmt, jobID)
				if err == nil {
					stats = appendStats(stats, auth0Stats)
				}
			}
		}
		return stats, errs
	}

	return nil, nil
}

func (s *Auth0Plugin) waitJob(jobID string) error {
	for {
		j, err := s.mgmt.Job.Read(jobID)
		if err != nil {
			return err
		}

		switch *j.Status {
		case "pending":
			time.Sleep(1 * time.Second)
			continue
		case "failed":
			return fmt.Errorf("job %s failed", jobID)
		case "completed":
			return nil
		default:
			return fmt.Errorf("unknown status")
		}
	}
}

func (s *Auth0Plugin) startJob() error {
	job := &management.Job{
		ConnectionID:        auth0.String(s.connectionID),
		Upsert:              auth0.Bool(true),
		SendCompletionEmail: auth0.Bool(false),
		Users:               s.users,
	}
	s.wg.Add(1)
	defer s.wg.Done()
	err := s.mgmt.Job.ImportUsers(job)
	if err != nil {
		return err
	}
	s.jobs = append(s.jobs, *job)

	return nil
}

func structToMap(in interface{}) (map[string]interface{}, int64, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, 0, err
	}
	res := make(map[string]interface{})
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, 0, err
	}
	size := int64(len(data))
	return res, size, nil
}

func retrieveJobSummary(mngmt *management.Management, jobID string) (map[string]interface{}, error) {
	job := &management.Job{}
	req, err := mngmt.NewRequest("GET", mngmt.URI("jobs", jobID), job)

	if err != nil {
		return nil, err
	}

	res, err := mngmt.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("request failed, status code: %d", res.StatusCode)
	}

	body := make(map[string]interface{})

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusAccepted {
		defer res.Body.Close()

		err := json.NewDecoder(res.Body).Decode(&body)

		if err != nil {
			return nil, fmt.Errorf("decoding response payload failed: %w", err)
		}

		if len(body) == 0 || body["summary"] == nil {
			return nil, errors.New("response body doesn't contain summary")
		}

		stats, ok := body["summary"].(map[string]interface{})
		if ok {
			return stats, nil
		}
	}

	return body, nil
}

func appendStats(plStats *plugin.Stats, auth0Stats map[string]interface{}) *plugin.Stats {
	if len(auth0Stats) == 0 {
		return plStats
	}
	total, _ := auth0Stats["total"].(float64)
	failed, _ := auth0Stats["failed"].(float64)
	updated, _ := auth0Stats["updated"].(float64)
	inserted, _ := auth0Stats["inserted"].(float64)

	return &plugin.Stats{
		Received: plStats.Received + int32(total),
		Created:  plStats.Created + int32(inserted),
		Updated:  plStats.Updated + int32(updated),
		Errors:   plStats.Errors + int32(failed),
	}
}
