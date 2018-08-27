package main

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_checkAndSaveSeqNum_fail(t *testing.T) {
	// pass in invalid seqnum format
	_, err := checkAndSaveSeqNum(noopLogger, 0, "/non/existing/dir")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to save the sequence number`)
}

func Test_checkAndSaveSeqNum_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fp := filepath.Join(dir, "seqnum")
	defer os.RemoveAll(dir)

	// no sequence number, 0 comes in.
	shouldExit, err := checkAndSaveSeqNum(noopLogger, 0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(noopLogger, 0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(noopLogger, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(noopLogger, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(noopLogger, 1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(noopLogger, 0, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)
}

func Test_parseVersionString_fail(t *testing.T) {
	_, err := parseAndLogAgentVersion(noopLogger, "helloWorld.zip")
	require.NotNil(t, err)
}

func Test_parseVersionString_success(t *testing.T) {
	_, err := parseAndLogAgentVersion(noopLogger, "agent/DesiredStateConfiguration_1.0.0.zip")
	require.Nil(t, err)
}

func Test_parseAndCompareExtensionVersions(t *testing.T) {
	extensions := []string{
		"Microsoft.GuestConfiguration.Edp.ConfigurationForLinux-0.4.0",
		"Microsoft.GuestConfiguration.Edp.ConfigurationForLinux-2.5.1",
		"Microsoft.GuestConfiguration.Edp.ConfigurationForLinux-1.7.8"}
	x, err := parseAndCompareExtensionVersions(noopLogger, extensions)
	require.Equal(t, "Microsoft.GuestConfiguration.Edp.ConfigurationForLinux-0.4.0", x)

	extensions = []string{
		"Microsoft.GuestConfiguration.ConfigurationForLinux-0.4.0",
		"Microsoft.GuestConfiguration.ConfigurationForLinux-2.5.1",
		"Microsoft.GuestConfiguration.ConfigurationForLinux-1.7.8"}
	x, err = parseAndCompareExtensionVersions(noopLogger, extensions)
	require.Equal(t, "Microsoft.GuestConfiguration.ConfigurationForLinux-0.4.0", x)

	require.Nil(t, err)
}

func Test_runCmd_fail(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.NotNil(t, runCmd(noopLogger, "wrongCmd", dir, handlerSettings{}))
}

func Test_runCmd_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.Nil(t, runCmd(noopLogger, "date", dir, handlerSettings{
		publicSettings: publicSettings{CommandToExecute: "date"},
	}), "command should run successfully")

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")

	require.Nil(t, runCmd(noopLogger, "", dir, handlerSettings{}))

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")
}

func Test_runCmd_withTestFile(t *testing.T) {
	dir := filepath.Join(DataDir, "testing")
	_, err := unzipAgent(noopLogger, "../integration-test/testdata/testing/", "testing.zip", DataDir)
	if err != nil {
		t.Fatal(err)
	}

	// print files in directory
	_, err = ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = runCmd(noopLogger, "bash ./testing.sh", dir, handlerSettings{})

	require.Nil(t, err)
}

func Test_unzip_fail(t *testing.T) {
	_, err := unzipAgent(noopLogger, "", "", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `no such file or directory`)

	_, err = unzipAgent(noopLogger, "../integration-test/testdata/testing/", "hello", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to find zip file`)

	_, err = unzipAgent(noopLogger, "../integration-test/testdata/testing/", "testing-corrupt.zip", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to open zip`)
}

func Test_unzip_pass(t *testing.T) {
	dir := filepath.Join(DataDir, UnzipAgentDir)
	fileNames, err := unzipAgent(noopLogger, "../"+AgentZipDir, AgentName, dir)
	require.Nil(t, err)
	require.NotEmpty(t, fileNames)

	dir = filepath.Join(DataDir, UnzipAgentDir)
	fileNames, err = unzipAgent(noopLogger, "../"+AgentZipDir, AgentName, dir)
	require.Nil(t, err)
	require.NotEmpty(t, fileNames)

	Test_cleanUpTests(t)
}

func Test_cleanUpTests(t *testing.T) {
	// delete the testing directory
	// if it does not exist, this will do nothing
	files := [3]string{UnzipAgentDir, "testing", "__MACOSX"}

	for _, file := range files {
		os.RemoveAll(file)

		exists := true
		if _, err := os.Stat(UnzipAgentDir); os.IsNotExist(err) {
			exists = false
		}
		require.False(t, exists)
	}
}
