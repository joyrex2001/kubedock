package main

import (
	_ "embed"

	"github.com/joyrex2001/kubedock/cmd"
)

//go:embed README.md
var readme string

//go:embed LICENSE
var license string

func main() {
	cmd.README = readme
	cmd.LICENSE = license
	cmd.Execute()
}
