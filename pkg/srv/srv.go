package srv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/idp-plugin-sdk/plugin"
	multierror "github.com/hashicorp/go-multierror"
	"gopkg.in/auth0.v5"
	"gopkg.in/auth0.v5/management"
)

const (
	maxBatchSize = int64(500 * 1024)
)

type Auth0Plugin struct {
	Config       *Auth0Config
	mgmt         *management.Management
	page         int
	finishedRead bool
	totalSize    int64
	jobs         []management.Job
	users        []map[string]interface{}
	connectionID string
	wg           sync.WaitGroup
}

func NewAuth0Plugin() *Auth0Plugin {
	return &Auth0Plugin{
		Config: &Auth0Config{},
	}
}

func (s *Auth0Plugin) GetConfig() plugin.PluginConfig {
	return &Auth0Config{}
}

func (s *Auth0Plugin) Open(cfg plugin.PluginConfig) error {
	config, ok := cfg.(*Auth0Config)
	if !ok {
		return errors.New("invalid config")
	}
	s.Config = config
	s.page = 0
	s.finishedRead = false

	mgnt, err := management.New(
		config.Domain,
		management.WithClientCredentials(
			config.ClientID,
			config.ClientSecret,
		))

	if err != nil {
		return nil
	}

	s.mgmt = mgnt

	c, err := mgnt.Connection.ReadByName("Username-Password-Authentication")
	if err != nil {
		return err
	}

	s.connectionID = auth0.StringValue(c.ID)
	return nil
}

func (s *Auth0Plugin) Read() ([]*api.User, error) {
	if s.finishedRead {
		return nil, io.EOF
	}

	opts := management.Page(s.page)
	ul, err := s.mgmt.User.List(opts)
	if err != nil {
		return nil, err
	}

	var errs error
	var users []*api.User
	for _, u := range ul.Users {
		user, err := Transform(u)
		if err != nil {
			errs = multierror.Append(errs, err)
		}

		users = append(users, user)
	}
	if !ul.HasNext() {
		s.finishedRead = true
	}
	s.page++

	return users, errs
}

func (s *Auth0Plugin) Write(user *api.User) error {
	u, err := TransformToAuth0(user)
	if err != nil {
		return err
	}

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

func (s *Auth0Plugin) Close() error {
	if len(s.users) > 0 {
		err := s.startJob()

		if err != nil {
			return err
		}
	}

	var errs error
	for _, j := range s.jobs {
		jobID := auth0.StringValue(j.ID)
		err := s.waitJob(jobID)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return nil
}

func (s *Auth0Plugin) waitJob(jobID string) error {
	for {
		j, err := s.mgmt.Job.Read(jobID)
		if err != nil {
			return err
		}

		switch *j.Status {
		case "pending":
			{
				time.Sleep(1 * time.Second)
				continue
			}
		case "failed":
			return fmt.Errorf("Job %s failed", jobID)
		case "completed":
			return nil
		default:
			return fmt.Errorf("Unknown status")
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
