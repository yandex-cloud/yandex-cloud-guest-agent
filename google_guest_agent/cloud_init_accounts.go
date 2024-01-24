// Please note that the code below is modified by YANDEX LLC

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/GoogleCloudPlatform/guest-agent/metadata"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/guest-agent/google_guest_agent/cfg"
	"github.com/GoogleCloudPlatform/guest-agent/utils"
	"github.com/GoogleCloudPlatform/guest-logging-go/logger"
)

var (
	// this parameter is platform dependent, currently limited number of distros are supported
	cloudInitSudoUsersFile = "/etc/sudoers.d/90-cloud-init-users"
)

type cloudInitAccountsMgr struct{}

func usersEqual(first, second []metadata.User) bool {
	if len(first) != len(second) {
		return false
	}

	mFirst := make(map[string]metadata.User, len(first))
	for _, user := range first {
		mFirst[user.Name] = user
	}

	for _, secondUser := range second {
		if firstUser, ok := mFirst[secondUser.Name]; ok {
			if !utils.SortedEqual(firstUser.SudoRules, secondUser.SudoRules) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}
func (a *cloudInitAccountsMgr) Diff(ctx context.Context) (bool, error) {
	// If we've enabled or disabled OS Login.
	oldOslogin, _, _ := getOSLoginEnabled(oldMetadata)
	newOslogin, _, _ := getOSLoginEnabled(newMetadata)
	if oldOslogin != newOslogin {
		return true, nil
	}

	// if metadata is update update users
	if !usersEqual(oldMetadata.Instance.Attributes.UserData, newMetadata.Instance.Attributes.UserData) {
		return true, nil
	}

	return false, nil
}

func (a *cloudInitAccountsMgr) Timeout(ctx context.Context) (bool, error) {
	return false, nil
}

func (a *cloudInitAccountsMgr) Disabled(ctx context.Context) (bool, error) {
	config := cfg.Get()
	oslogin, _, _ := getOSLoginEnabled(newMetadata)
	return false || runtime.GOOS == "windows" || oslogin || !config.Daemons.AccountsDaemon, nil
}

func (a *cloudInitAccountsMgr) Set(ctx context.Context) error {
	// For cloud init users update sudoers file
	if newMetadata.Instance.Attributes.UserData != nil {
		logger.Infof("Creating cloud init sudoers.")
		for _, userData := range newMetadata.Instance.Attributes.UserData {
			err := writeCloudInitSudoRules(userData.Name, userData.SudoRules, cloudInitSudoUsersFile, clockInstance)
			if err != nil {
				logger.Errorf("Error updating cloud init sudo rules for %s: %v.", userData.Name, err)
			}
		}
	}

	// no keys are provided then os-login is enable and sudo file for cloud-init users must be removed
	if newMetadata.Instance.Attributes.SSHKeys == nil {
		logger.Infof("Removing cloud init sudoers.")
		if err := os.Remove(cloudInitSudoUsersFile); err != nil {
			logger.Errorf("Error removing cloud init sudoers: %v.", err)
		}
	}

	// Start SSHD if not started. We do this in agent instead of adding a
	// Wants= directive, and here instead of instance setup, so that this
	// can be disabled by the instance configs file.
	for _, svc := range []string{"ssh", "sshd"} {
		// Ignore output, it's just a best effort.
		systemctlStart(ctx, svc)
	}

	return nil
}

// writeCloudInitSudoRules restores sudo rules for cloud-init users
func writeCloudInitSudoRules(user string, rules []string, sudoersFile string, clock Clock) error {
	lines := []string{
		"",
		fmt.Sprintf("# User rules for %s", user),
	}

	if rules != nil && len(rules) > 0 {
		for _, rule := range rules {
			lines = append(lines, fmt.Sprintf("%s %s", user, rule))
		}
	}

	content := strings.Join(lines, "\n")
	content = content + "\n"

	if _, err := os.Stat(sudoersFile); err == nil {
		fileContent, err := utils.ReadTextFile(sudoersFile)
		if err != nil {
			return err
		}

		if !strings.Contains(fileContent, content) {
			return utils.AppendFile([]byte(content), sudoersFile)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		err = utils.WriteFile([]byte(makeCloudInitSudoersHeader(clock)+"\n"+content), sudoersFile, 0440)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// makeCloudInitSudoersHeader generates information header for a new sudo file generated from cloud-init configuration
func makeCloudInitSudoersHeader(clock Clock) string {
	now := clock.Now().UTC()
	formattedNow := now.Format(time.RFC1123Z)
	return fmt.Sprintf("# Created by guest-agent (for cloud-init) v. %s on %s", version, formattedNow)
}
