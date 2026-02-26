package stages

import (
	"fmt"
	"time"

	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func creditTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "credit",
		Timeout:       30 * time.Second,
		TestFunc:      testCredit,
		RequiredFiles: []string{"credit.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "credit.c",
			Output:           "credit",
			IncludeParentDir: true,
		},
	}
}

func testCredit(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例 (对齐 CS50 check50)
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// AMEX (34 或 37 开头, 15位)
		{"378282246310005", "AMEX", "identifies 378282246310005 as AMEX"},
		{"371449635398431", "AMEX", "identifies 371449635398431 as AMEX"},
		// MASTERCARD (51-55 开头, 16位)
		{"5555555555554444", "MASTERCARD", "identifies 5555555555554444 as MASTERCARD"},
		{"5105105105105100", "MASTERCARD", "identifies 5105105105105100 as MASTERCARD"},
		// VISA (4 开头, 13或16位)
		{"4111111111111111", "VISA", "identifies 4111111111111111 as VISA"},
		{"4012888888881881", "VISA", "identifies 4012888888881881 as VISA"},
		{"4222222222222", "VISA", "identifies 4222222222222 as VISA"},
		// INVALID - 各种无效情况
		{"1234567890", "INVALID", "identifies 1234567890 as INVALID"},
		{"369421438430814", "INVALID", "identifies 369421438430814 as INVALID"},
		{"4062901840", "INVALID", "identifies 4062901840 as INVALID"},
		{"5673598276138003", "INVALID", "identifies 5673598276138003 as INVALID"},
		{"4111111111111113", "INVALID", "identifies 4111111111111113 as INVALID"},
		{"4222222222223", "INVALID", "identifies 4222222222223 as INVALID"},
		{"3400000000000620", "INVALID", "identifies 3400000000000620 as INVALID"},
		{"430000000000000", "INVALID", "identifies 430000000000000 as INVALID"},
	}

	for _, tc := range tests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "credit").
			WithTimeout(5 * time.Second).
			Stdin(tc.input).
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	logger.Successf("All credit tests passed!")
	return nil
}
