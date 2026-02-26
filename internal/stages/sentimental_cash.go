package stages

import (
	"fmt"
	"time"

	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func sentimentalCashTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "sentimental-cash",
		Timeout:       30 * time.Second,
		TestFunc:      testSentimentalCash,
		RequiredFiles: []string{"cash.py"},
	}
}

func testSentimentalCash(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试有效输入（对齐 CS50 check50，使用浮点数美元）
	validTests := []struct {
		input    string
		expected string
		name     string
	}{
		{"0.41", "4", "input of 0.41 yields output of 4"},
		{"0.01", "1", "input of 0.01 yields output of 1"},
		{"0.15", "2", "input of 0.15 yields output of 2"},
		{"1.6", "7", "input of 1.6 yields output of 7"},
		{"23", "92", "input of 23 yields output of 92"},
		{"4.2", "18", "input of 4.2 yields output of 18"},
	}

	for _, tc := range validTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "python3", "cash.py").
			WithTimeout(5 * time.Second).
			Stdin(tc.input).
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 3. 测试拒绝无效输入 (对齐 CS50 check50)
	rejectTests := []struct {
		input string
		name  string
	}{
		{"-1", "rejects a negative input like -1"},
		{"foo", "rejects a non-numeric input of \"foo\""},
		{"", "rejects a non-numeric input of \"\""},
	}

	for _, tc := range rejectTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "python3", "cash.py").
			WithTimeout(5 * time.Second).
			WithPty().
			Start().
			SendLine(tc.input).
			Reject(200 * time.Millisecond)

		if err := r.Error(); err != nil {
			r.Kill()
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		r.Kill()
		logger.Successf("✓ %s", tc.name)
	}

	logger.Successf("All tests passed!")
	return nil
}
