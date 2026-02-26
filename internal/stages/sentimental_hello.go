package stages

import (
	"fmt"
	"time"

	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func sentimentalHelloTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "sentimental-hello",
		Timeout:       30 * time.Second,
		TestFunc:      testSentimentalHello,
		RequiredFiles: []string{"hello.py"},
	}
}

func testSentimentalHello(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例：对齐 CS50 check50 官方测试
	testCases := []struct {
		name     string
		expected string
	}{
		{"David", "hello, David"},
		{"Veronica", "hello, Veronica"},
		{"Brian", "hello, Brian"},
	}

	for _, tc := range testCases {
		logger.Infof("Testing with input %q...", tc.name)

		r := runner.Run(workDir, "python3", "hello.py").
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
