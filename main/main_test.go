package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseCmd_success(t *testing.T) {
	strs := []string{"install", "enable", "disable", "uninstall", "update",
		"INstaLL", "enAble", "disABLE", "UNINSTALL", "uPdAtE"}
	for _, c := range strs {
		cmd := parseCmd([]string{c})
		_, ok := cmds[cmd.name]
		assert.Equal(t, ok, true, "Incorrect command")
	}
}

//
//func Test_parseCmd_failure(t *testing.T) {
//	strs := []string{"install", "enable", "disable", "uninstall", "update",
//		"INstaLL", "enAble", "disABLE", "UNINSTALL", "uPdAtE"}
//	for _, c := range strs {
//		cmd := parseCmd([]string{c})
//		_, ok := cmds[cmd.name]
//		assert.Equal(t, ok, true, "Incorrect command")
//	}
//
//	args := []string{}
//	cmd := parseCmd(args)
//	_, ok := cmds[cmd.name]
//	assert.Equal(t, ok, true, "Incorrect command")
//}
//
