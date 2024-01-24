// Please note that the code below is modified by YANDEX LLC

package metadata

import (
	"encoding/json"

	"github.com/GoogleCloudPlatform/guest-logging-go/logger"
	"gopkg.in/yaml.v3"
)

// User describes the User metadata keys.
type User struct {
	Name      string
	SshKeys   []string
	SudoRules []string
}

// UserData is a slice of User.
type UserData []User

// UnmarshalJSON unmarshals b into UserData.
func (u *UserData) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	type inner struct {
		Users UserData `yaml:"users"`
	}
	var tempUserData inner
	if err := yaml.Unmarshal([]byte(s), &tempUserData); err != nil {
		logger.Infof("User-data yaml is invalid. Error: %+v", err)
		// ignore invalid user-data
		return nil
	}

	for _, userData := range tempUserData.Users {
		if userData.Name == "" {
			continue
		}

		*u = append(*u, User{
			Name:      userData.Name,
			SshKeys:   userData.SshKeys,
			SudoRules: userData.SudoRules,
		})
	}

	return nil

}

func (s *User) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type innerUser struct {
		Name      string      `yaml:"name"`
		Snapuser  string      `yaml:"snapuser"`
		SshKeys   []string    `yaml:"ssh_authorized_keys"`
		SudoRules interface{} `yaml:"sudo"`
	}
	var tempUser innerUser
	if err := unmarshal(&tempUser); err != nil {
		logger.Infof("User-data user will not be parsed. Error: %+v", err)
		// ignore line if the yaml invalid or the user is a "default"
		return nil
	}

	if tempUser.Snapuser != "" {
		logger.Infof("User-data snap user is ignored")
		return nil
	}

	sudoRules := make([]string, 0)
	if str, ok := tempUser.SudoRules.(string); ok {
		sudoRules = append(sudoRules, str)
	} else if rules, ok := tempUser.SudoRules.([]interface{}); ok {
		for _, v := range rules {
			if rule, ok := v.(string); ok {
				sudoRules = append(sudoRules, rule)
			} else {
				logger.Errorf("Cannot parse cloud-init sudo rule for user %s.", tempUser.Name)
			}
		}
	} else if _, ok := tempUser.SudoRules.(bool); !ok {
		logger.Errorf("Cannot parse sudo rules")
	}

	s.Name = tempUser.Name
	s.SshKeys = tempUser.SshKeys
	s.SudoRules = sudoRules
	return nil
}
