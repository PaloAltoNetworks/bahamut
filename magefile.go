// +build mage

// nolint
package main

import (
	"github.com/aporeto-inc/domingo/golang"
	"github.com/magefile/mage/mg"
)

func init() {
	domingo.SetProjectName("bahamut")
}

// Init initialize the project.
func Init() {
	mg.Deps(
		domingo.Init,
	)
}

// Test runs unit tests.
func Test() {
	mg.Deps(
		domingo.Lint,
		domingo.Test,
	)
}
