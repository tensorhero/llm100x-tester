package main

import (
	"os"

	"github.com/hellobyte-dev/llm100x-tester/internal/stages"
	tester_utils "github.com/hellobyte-dev/tester-utils"
)

func main() {
	definition := stages.GetDefinition()
	os.Exit(tester_utils.Run(os.Args[1:], definition))
}
