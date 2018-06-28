package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_commandsExist(t *testing.T) {
	expect := []string{"install", "enable", "disable", "uninstall", "update"}
	for _, c := range expect {
		_, ok := cmds[c]
		if !ok {
			t.Fatalf("cmd '%s' is not handled", c)
		}
	}
}

func Test_parseCmd_success(t *testing.T) {
	strs := []string{"install", "enable", "disable", "uninstall", "update",
		"INstaLL", "enAble", "disABLE", "UNINSTALL", "uPdAtE"}
	for _, c := range strs {
		cmd := parseCmd([]string{c})
		_, ok := cmds[cmd.name]
		assert.Equal(t, ok, true, "Incorrect command")
	}
}
