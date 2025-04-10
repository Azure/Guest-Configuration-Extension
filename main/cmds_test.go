package main

import (
	"testing"

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

func Test_install(t *testing.T) {
	err := install(ExtensionLogger{newNoopLogger()},
		vmextension.HandlerEnvironment{},
		0)
	require.Nil(t, err)
}

func Test_enablePre(t *testing.T) {
	dir := filepath.Join(MostRecentSequence)
	os.RemoveAll(dir)
	defer os.RemoveAll(dir)

	err := enablePre(noopLogger, 0)
	require.Nil(t, err)

	err = enablePre(noopLogger, 0)
	require.Nil(t, err)

	err = enablePre(noopLogger, 1)
	require.Nil(t, err)

	err = enablePre(noopLogger, 4)
	require.Nil(t, err)
}
