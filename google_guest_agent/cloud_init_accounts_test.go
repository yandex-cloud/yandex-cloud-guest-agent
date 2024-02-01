// Please note that the code below is modified by YANDEX LLC

package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/guest-agent/metadata"
	"github.com/GoogleCloudPlatform/guest-agent/utils"
)

var (
	cloudInitTestSudoers = []string{"ALL=(ALL) NOPASSWD:ALL"}
	testTime             = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
)

const (
	cloudInitTestUser  = "cloud-init-test-user"
	cloudInitTestUser2 = "cloud-init-test-user-2"
	testVersion        = "test-version"
)

func TestUserDiff(t *testing.T) {
	testCases := []struct {
		userListFirst  []metadata.User
		userListSecond []metadata.User
		same           bool
	}{
		{
			userListFirst: []metadata.User{
				{
					Name:      "user-1",
					SudoRules: []string{"rule2", "rule1"},
				},
				{
					Name:      "user-2",
					SudoRules: []string{"rule2"},
				},
			},
			userListSecond: []metadata.User{
				{
					Name:      "user-2",
					SudoRules: []string{"rule2"},
				},
				{
					Name:      "user-1",
					SudoRules: []string{"rule1", "rule2"},
				},
			},
			same: true,
		},
		{
			userListFirst: []metadata.User{
				{
					Name:      "user-1",
					SudoRules: []string{"rule1"},
				},
				{
					Name:      "user-2",
					SudoRules: []string{"rule2", "rule1"},
				},
			},
			userListSecond: []metadata.User{
				{
					Name:      "user-2",
					SudoRules: []string{"rule1", "rule3"},
				},
				{
					Name:      "user-1",
					SudoRules: []string{"rule1"},
				},
			},
			same: false,
		},
	}

	for _, test := range testCases {
		if usersEqual(test.userListFirst, test.userListSecond) != test.same {
			t.Fatalf(`Incorrect test result`)
		}
	}
}

func TestWriteCloudInitSudoRules(t *testing.T) {
	version = testVersion
	sudoersFile := filepath.Join(t.TempDir(), "90-cloud-init-users")

	var expectedResult = `# Created by guest-agent (for cloud-init) v. test-version on Tue, 17 Nov 2009 20:34:58 +0000

# User rules for cloud-init-test-user
cloud-init-test-user ALL=(ALL) NOPASSWD:ALL
`

	err := writeCloudInitSudoRules(cloudInitTestUser, cloudInitTestSudoers, sudoersFile, newMockClock(testTime))
	if err != nil {
		t.Errorf("Failed to write cloud init sudoers: %v", err)
		return
	}
	actualResult, err := utils.ReadTextFile(sudoersFile)
	if err != nil {
		t.Errorf("Failed to read sudoers: %v", err)
		return
	}
	if expectedResult != actualResult {
		t.Fatalf(`Sudoers file is incorrect`)
	}
}

func TestAppendCloudInitSudoRules(t *testing.T) {
	version = testVersion
	sudoersFile := filepath.Join(t.TempDir(), "90-cloud-init-users")
	var initialSudoersContent = `# Created by guest-agent (for cloud-init) v. test-version on Tue, 17 Nov 2009 20:34:58 +0000

# User rules for cloud-init-test-user
cloud-init-test-user ALL=(ALL) NOPASSWD:ALL
`
	var expectedResult = `# Created by guest-agent (for cloud-init) v. test-version on Tue, 17 Nov 2009 20:34:58 +0000

# User rules for cloud-init-test-user
cloud-init-test-user ALL=(ALL) NOPASSWD:ALL

# User rules for cloud-init-test-user-2
cloud-init-test-user-2 ALL=(ALL) NOPASSWD:ALL
`

	err := utils.WriteFile([]byte(initialSudoersContent), sudoersFile, 0777)
	if err != nil {
		t.Errorf("Cannot initialized sudoers file: %v", err)
		return
	}

	for _, user := range []string{cloudInitTestUser2, cloudInitTestUser} {
		err = writeCloudInitSudoRules(user, cloudInitTestSudoers, sudoersFile, newMockClock(testTime))
		if err != nil {
			t.Errorf("Failed to write cloud init sudoers: %v", err)
			return
		}
	}

	actualResult, err := utils.ReadTextFile(sudoersFile)
	if err != nil {
		t.Errorf("Failed to read sudoers: %v", err)
		return
	}
	if expectedResult != actualResult {
		t.Fatalf(`Sudoers file is incorrect`)
	}
}
