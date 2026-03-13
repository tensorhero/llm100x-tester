package stages

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func fiftyvilleTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "fiftyville",
		Timeout:       60 * time.Second,
		TestFunc:      testFiftyville,
		RequiredFiles: []string{"log.sql", "answers.txt"},
	}
}

func testFiftyville(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 检查 log.sql 包含 SELECT 查询
	logger.Infof("Checking log file contains SELECT queries...")
	logContent, err := os.ReadFile(filepath.Join(workDir, "log.sql"))
	if err != nil {
		return fmt.Errorf("failed to read log.sql: %v", err)
	}
	logLower := strings.ToLower(string(logContent))
	if !strings.Contains(logLower, "select") {
		return fmt.Errorf("missing SELECT queries in log.sql")
	}
	logger.Successf("log file contains SELECT queries")

	// 3. 检查谜题是否解决
	logger.Infof("Checking mystery solved...")
	answersContent, err := os.ReadFile(filepath.Join(workDir, "answers.txt"))
	if err != nil {
		return fmt.Errorf("failed to read answers.txt: %v", err)
	}
	answersLower := strings.ToLower(string(answersContent))

	// 答案 (与 CS50 check50 对齐)
	// thief: bruce (hex: 6272756365)
	// city: new york (hex: 6e657720796f726b)
	// accomplice: robin (hex: 726f62696e)
	thief := "bruce"
	city := "new york"
	accomplice := "robin"

	// 检查格式 - 每个关键词只能出现一次
	for _, q := range []string{"thief is", "escaped to", "accomplice is"} {
		if strings.Count(answersLower, q) > 1 {
			return fmt.Errorf("invalid answers.txt formatting: '%s' appears more than once", q)
		}
	}

	// 使用正则匹配答案
	thiefPattern := regexp.MustCompile(`thief\s*is\s*:?\s*` + regexp.QuoteMeta(thief))
	cityPattern := regexp.MustCompile(`escaped\s*to\s*:?\s*` + regexp.QuoteMeta(city))
	accomplicePattern := regexp.MustCompile(`accomplice\s*is\s*:?\s*` + regexp.QuoteMeta(accomplice))

	if !thiefPattern.MatchString(answersLower) {
		return fmt.Errorf("answers.txt does not correctly identify the thief")
	}
	if !cityPattern.MatchString(answersLower) {
		return fmt.Errorf("answers.txt does not correctly identify the city the thief escaped to")
	}
	if !accomplicePattern.MatchString(answersLower) {
		return fmt.Errorf("answers.txt does not correctly identify the accomplice")
	}

	logger.Successf("mystery solved")
	logger.Successf("All tests passed!")
	return nil
}
