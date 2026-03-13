package stages

import (
	"fmt"
	"time"

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func helloTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "hello",
		Timeout:       30 * time.Second,
		TestFunc:      testHello,
		RequiredFiles: []string{"hello.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "hello.c",
			Output:           "hello",
			IncludeParentDir: true,
		},
	}
}

func testHello(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例：对齐 CS50 check50 官方测试
	testCases := []struct {
		name     string
		expected string
	}{
		{"Emma", "Emma"},
		{"Rodrigo", "Rodrigo"},
	}

	for _, tc := range testCases {
		logger.Infof("Testing with input %q...", tc.name)

		r := runner.Run(workDir, "hello").
			WithTimeout(5 * time.Second).
			Stdin(tc.name).
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for input %q: %v", tc.name, err)
		}

		logger.Successf("✓ Output correct for input %q", tc.name)
	}

	logger.Successf("All tests passed!")
	return nil
}
