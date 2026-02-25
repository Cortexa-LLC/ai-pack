package main

import "github.com/cortexa-llc/ai-pack/cmd/agent/commands"

// Version info injected at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	commands.Execute(Version, Commit, BuildTime)
}
