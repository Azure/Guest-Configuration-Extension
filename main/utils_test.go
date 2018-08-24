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
	_, err := checkAndSaveSeqNum(0, "/non/existing/dir")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to save the sequence number`)
}

func Test_checkAndSaveSeqNum_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	fp := filepath.Join(dir, "seqnum")
	defer os.RemoveAll(dir)

	// no sequence number, 0 comes in.
	shouldExit, err := checkAndSaveSeqNum(0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(0, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=0, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=1 comes in.
	shouldExit, err = checkAndSaveSeqNum(1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in.
	shouldExit, err = checkAndSaveSeqNum(1, fp)
	require.Nil(t, err)
	require.False(t, shouldExit)

	// file=1, seq=0 comes in. (should exit)
	shouldExit, err = checkAndSaveSeqNum(0, fp)
	require.Nil(t, err)
	require.True(t, shouldExit)
}

func Test_parseVersionString_fail(t *testing.T) {
	_, err := parseAndLogAgentVersion("helloWorld.zip")
	require.NotNil(t, err)
}

func Test_parseVersionString_success(t *testing.T) {
	_, err := parseAndLogAgentVersion(agentZip)
	require.Nil(t, err)
}

func Test_getOldAgentPath(t *testing.T) {
	_, err := getOldAgentPath()
	require.Nil(t, err)
}

func Test_runCmd_fail(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.NotNil(t, runCmd("wrongCmd", dir, handlerSettings{}))
}

func Test_runCmd_success(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	require.Nil(t, runCmd("date", dir, handlerSettings{
		publicSettings: publicSettings{CommandToExecute: "date"},
	}), "command should run successfully")

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")

	require.Nil(t, runCmd("", dir, handlerSettings{}))

	// check stdout stderr files
	_, err = os.Stat(filepath.Join(dir, "stdout"))
	require.Nil(t, err, "stdout should exist")
	_, err = os.Stat(filepath.Join(dir, "stderr"))
	require.Nil(t, err, "stderr should exist")
}

func Test_runCmd_withTestFile(t *testing.T) {
	dir := filepath.Join(dataDir, "testing")
	_, err := unzip("../testing/testing.zip", dataDir)

	// print files in directory
	_, err = ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = runCmd("bash ./testing.sh", dir, handlerSettings{})

	require.Nil(t, err)
}

func Test_unzip_fail(t *testing.T) {
	_, err := unzip("", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to open zip`)

	_, err = unzip("hello.zip", "agent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `failed to open zip`)
}

func Test_unzip_pass(t *testing.T) {
	dir := filepath.Join(dataDir, agentDir)
	filenames, err := unzip("../"+agentZip, dir)
	require.Nil(t, err)
	require.NotEmpty(t, filenames)

	dir = filepath.Join(dataDir, agentDir)
	filenames, err = unzip("../"+agentZip, dir)
	require.Nil(t, err)
	require.NotEmpty(t, filenames)

	Test_cleanUpTests(t)
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
