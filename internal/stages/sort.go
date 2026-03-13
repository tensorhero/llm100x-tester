package stages

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func sortTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "sort",
		Timeout:       30 * time.Second,
		TestFunc:      testSort,
		RequiredFiles: []string{"answers.txt"},
	}
}

func testSort(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger

	// 读取 answers.txt 内容
	content, err := harness.ReadFile("answers.txt")
	if err != nil {
		return fmt.Errorf("could not read answers.txt: %v", err)
	}
	answers := string(content)

	// 3. 检查是否还有未回答的问题
	logger.Infof("Checking all questions are answered...")
	if strings.Contains(answers, "TODO") {
		return fmt.Errorf("not all questions answered - still contains TODO")
	}
	logger.Successf("all questions answered")

	// 4. 检查排序算法识别是否正确
	logger.Infof("Checking that sorts are classified correctly...")

	// CS50 check50 的正确答案
	expectedPatterns := []struct {
		pattern string
		desc    string
	}{
		{`sort1 uses:\s*[Bb][Uu][Bb][Bb][Ll][Ee]`, "sort1 should use Bubble sort"},
		{`sort2 uses:\s*[Mm][Ee][Rr][Gg][Ee]`, "sort2 should use Merge sort"},
		{`sort3 uses:\s*[Ss][Ee][Ll][Ee][Cc][Tt][Ii][Oo][Nn]`, "sort3 should use Selection sort"},
	}

	for _, ep := range expectedPatterns {
		re := regexp.MustCompile(ep.pattern)
		if !re.MatchString(answers) {
			return fmt.Errorf("incorrect assignment of sorts: %s", ep.desc)
		}
		logger.Successf("✓ %s", ep.desc)
	}

	logger.Successf("All sort tests passed!")
	return nil
}
