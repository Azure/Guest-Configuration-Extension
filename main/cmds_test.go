package main

import (
	"testing"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/stretchr/testify/require"
)

func Test_commandsExist(t *testing.T) {
	// we expect these subcommands to be handled
	expect := []string{"install", "enable", "disable", "uninstall", "update"}
	for _, c := range expect {
		_, ok := cmds[c]
		if !ok {
			t.Fatalf("cmd '%s' is not handled", c)
		}
	}
}

func Test_commands_shouldReportStatus(t *testing.T) {
	// - certain extension invocations are supposed to write 'N.status' files and some do not.

	// these subcommands should NOT report status
	require.False(t, cmds["install"].shouldReportStatus, "install should not report status")
	require.False(t, cmds["uninstall"].shouldReportStatus, "uninstall should not report status")

	// these subcommands SHOULD report status
	require.True(t, cmds["enable"].shouldReportStatus, "enable should report status")
	require.True(t, cmds["disable"].shouldReportStatus, "disable should report status")
	require.True(t, cmds["update"].shouldReportStatus, "update should report status")
}

func Test_checkAndSaveSeqNum_fail(t *testing.T) {
	// pass in invalid seqnum format
	_, err := checkAndSaveSeqNum(newLogger(), 0, "/non/existing/dir")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to save the sequence number`)
}

func Test_checkAndSaveSeqNum_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fp := filepath.Join(dir, "seqnum")
	defer os.RemoveAll(dir)

	nop := newLogger()

	// no sequence number, 0 comes in.
	shouldExit, err := checkAndSaveSeqNum(nop, 0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(nop, 0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(nop, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(nop, 0, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)
}

func Test_parseVersionString_fail(t *testing.T) {
	_, err := parseVersionString("helloWorld.zip")
	require.NotNil(t, err)
}

func Test_parseVersionString_success(t *testing.T) {
	_, err := parseVersionString(agentZip)
	require.Nil(t, err)
}

func Test_runCmd_fail(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.NotNil(t, runCmd(newLogger(), "wrongCmd", dir, handlerSettings{}))
}

func Test_runCmd_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.Nil(t, runCmd(newLogger(), "date", dir, handlerSettings{
		publicSettings: publicSettings{CommandToExecute: "date"},
	}), "command should run successfully")

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")

	require.Nil(t, runCmd(newLogger(), "", dir, handlerSettings{}))

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")
}

func Test_runCmd_withTestFile(t *testing.T) {
	dir := filepath.Join(dataDir, "testing")
	_, err := unzip(newLogger(), "../testing/testing.zip", dataDir)

	// print files in directory
	_, err = ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = runCmd(newLogger(), "bash ./testing.sh", dir, handlerSettings{})

	require.Nil(t, err)
}

func Test_unzip_fail(t *testing.T) {
	_, err := unzip(newLogger(), "", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to open zip`)

	_, err = unzip(newLogger(), "hello.zip", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to open zip`)
}

func Test_unzip_pass(t *testing.T) {
	dir := filepath.Join(dataDir, agentDir)
	filenames, err := unzip(newLogger(), "../"+agentZip, dir)
	require.Nil(t, err)
	require.NotEmpty(t, filenames)

	dir = filepath.Join(dataDir, agentDir)
	filenames, err = unzip(newLogger(), "../"+agentZip, dir)
	require.Nil(t, err)
	require.NotEmpty(t, filenames)

	Test_cleanUpTests(t)
}

func Test_install(t *testing.T) {
	message, err := install(newLogger(), vmextension.HandlerEnvironment{}, 0)
	require.Nil(t, err)
	require.Empty(t, message)
}

func Test_enablePre(t *testing.T) {
	dir := filepath.Join(mostRecentSequence)
	os.RemoveAll(dir)
	nop := newLogger()
	defer os.RemoveAll(dir)

	err := enablePre(nop, 0)
	require.Nil(t, err)

	err = enablePre(nop, 0)
	require.Nil(t, err)

	err = enablePre(nop, 1)
	require.Nil(t, err)

	err = enablePre(nop, 4)
	require.Nil(t, err)
}

func Test_cleanUpTests(t *testing.T) {
	// delete the testing directory
	// if it does not exist, this will do nothing
	files := [3]string{agentDir, "testing", "__MACOSX"}

	for _, file := range files {
		os.RemoveAll(file)

		exists := true
		if _, err := os.Stat(agentDir); os.IsNotExist(err) {
			exists = false
		}
		require.False(t, exists)
	}
}
