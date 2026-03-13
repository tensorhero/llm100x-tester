package main

import (
	"os"

	"github.com/tensorhero/llm100x-tester/internal/stages"
	tester_utils "github.com/tensorhero/tester-utils"
)

func main() {
	definition := stages.GetDefinition()
	os.Exit(tester_utils.Run(os.Args[1:], definition))
}
