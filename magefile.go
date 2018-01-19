// +build mage

// nolint
package main

import (
	"github.com/aporeto-inc/domingo/golang"
	"github.com/magefile/mage/mg"
)

// Init initialize the project.
func Init() {
	mg.Deps(
		domingo.InstallDependencies,
	)
}

// Test runs the unit tests.
func Test() {
	mg.Deps(
		domingo.Lint,
		domingo.Test,
	)
}
