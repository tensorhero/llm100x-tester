package stages

import (
	"fmt"
	"time"

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func dnaTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "dna",
		Timeout:       60 * time.Second,
		TestFunc:      testDna,
		RequiredFiles: []string{"dna.py"},
	}
}

func testDna(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例 (对齐 CS50 check50 的 test1-test20)
	tests := []struct {
		database string
		sequence string
		expected string
		name     string
	}{
		// Small database tests (1-4)
		{"databases/small.csv", "sequences/1.txt", "Bob", "correctly identifies sequences/1.txt"},
		{"databases/small.csv", "sequences/2.txt", "No match", "correctly identifies sequences/2.txt"},
		{"databases/small.csv", "sequences/3.txt", "No match", "correctly identifies sequences/3.txt"},
		{"databases/small.csv", "sequences/4.txt", "Alice", "correctly identifies sequences/4.txt"},
		// Large database tests (5-20)
		{"databases/large.csv", "sequences/5.txt", "Lavender", "correctly identifies sequences/5.txt"},
		{"databases/large.csv", "sequences/6.txt", "Luna", "correctly identifies sequences/6.txt"},
		{"databases/large.csv", "sequences/7.txt", "Ron", "correctly identifies sequences/7.txt"},
		{"databases/large.csv", "sequences/8.txt", "Ginny", "correctly identifies sequences/8.txt"},
		{"databases/large.csv", "sequences/9.txt", "Draco", "correctly identifies sequences/9.txt"},
		{"databases/large.csv", "sequences/10.txt", "Albus", "correctly identifies sequences/10.txt"},
		{"databases/large.csv", "sequences/11.txt", "Hermione", "correctly identifies sequences/11.txt"},
		{"databases/large.csv", "sequences/12.txt", "Lily", "correctly identifies sequences/12.txt"},
		{"databases/large.csv", "sequences/13.txt", "No match", "correctly identifies sequences/13.txt"},
		{"databases/large.csv", "sequences/14.txt", "Severus", "correctly identifies sequences/14.txt"},
		{"databases/large.csv", "sequences/15.txt", "Sirius", "correctly identifies sequences/15.txt"},
		{"databases/large.csv", "sequences/16.txt", "No match", "correctly identifies sequences/16.txt"},
		{"databases/large.csv", "sequences/17.txt", "Harry", "correctly identifies sequences/17.txt"},
		{"databases/large.csv", "sequences/18.txt", "No match", "correctly identifies sequences/18.txt"},
		{"databases/large.csv", "sequences/19.txt", "Fred", "correctly identifies sequences/19.txt"},
		{"databases/large.csv", "sequences/20.txt", "No match", "correctly identifies sequences/20.txt"},
	}

	for _, tc := range tests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "python3", "dna.py", tc.database, tc.sequence).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	logger.Successf("All tests passed!")
	return nil
}
