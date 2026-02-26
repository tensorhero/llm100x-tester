package stages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func pluralityTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "plurality",
		Timeout:       30 * time.Second,
		TestFunc:      testPlurality,
		RequiredFiles: []string{"plurality.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "plurality.c",
			Output:           "plurality",
			IncludeParentDir: true,
		},
	}
}

func testPlurality(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 构建测试套件
	// 读取学生的 plurality.c，将 main 重命名为 distro_main
	pluralityCode, err := harness.ReadFile("plurality.c")
	if err != nil {
		return fmt.Errorf("could not read plurality.c: %v", err)
	}

	// 使用正则替换 main 函数
	mainRegex := regexp.MustCompile(`int\s+main\s*\(`)
	modifiedCode := mainRegex.ReplaceAllString(string(pluralityCode), "int distro_main(")

	// 从学生目录读取测试代码
	testCodeBytes, err := harness.ReadFile("plurality_test.c")
	if err != nil {
		return fmt.Errorf("plurality_test.c does not exist: %v", err)
	}
	testCode := string(testCodeBytes)

	// 写入组合的测试文件
	combinedCode := modifiedCode + "\n" + testCode
	testFilePath := filepath.Join(workDir, "plurality_combined_test.c")
	if err := os.WriteFile(testFilePath, []byte(combinedCode), 0644); err != nil {
		return fmt.Errorf("could not write test file: %v", err)
	}

	// 编译测试程序
	logger.Infof("Compiling test harness...")
	cmd := exec.Command("clang", "-o", "plurality_test", "plurality_combined_test.c", "-I..", "-lm", "-Wall")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("test harness does not compile: %s\n%s", err, string(out))
	}
	logger.Successf("test harness compiles")

	// 4. 运行 vote 函数测试
	voteTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"0", "0", "true", "vote returns true when given name of first candidate"},
		{"0", "1", "true", "vote returns true when given name of middle candidate"},
		{"0", "2", "true", "vote returns true when given name of last candidate"},
		{"0", "3", "false", "vote returns false when given name of invalid candidate"},
		{"0", "4", "1 0 0", "vote produces correct counts when all votes are zero"},
		{"0", "5", "2 8 0", "vote produces correct counts after some have already voted"},
		{"0", "6", "2 8 0", "vote leaves vote counts unchanged when voting for invalid candidate"},
	}

	for _, tc := range voteTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "plurality_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 5. 运行 print_winner 函数测试
	winnerTests := []struct {
		setup    string
		test     string
		expected []string // 期望的获胜者（可能多个）
		name     string
	}{
		{"0", "7", []string{"Alice"}, "print_winner identifies Alice as winner of election"},
		{"0", "8", []string{"Bob"}, "print_winner identifies Bob as winner of election"},
		{"0", "9", []string{"Charlie"}, "print_winner identifies Charlie as winner of election"},
		{"0", "10", []string{"Alice", "Bob"}, "print_winner prints multiple winners in case of tie"},
		{"0", "11", []string{"Alice", "Bob", "Charlie"}, "print_winner prints all names when all candidates are tied"},
	}

	for _, tc := range winnerTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "plurality_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		// 检查输出是否包含所有期望的获胜者
		output := r.GetStdout()
		actualWinners := parseWinners(output)

		if !winnersMatch(tc.expected, actualWinners) {
			return fmt.Errorf("test failed for %s: expected winners %v, got %v", tc.name, tc.expected, actualWinners)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 清理测试文件
	os.Remove(testFilePath)
	os.Remove(filepath.Join(workDir, "plurality_test"))

	logger.Successf("All plurality tests passed!")
	return nil
}

// parseWinners 从输出中解析获胜者名单
func parseWinners(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	winners := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			winners = append(winners, line)
		}
	}
	return winners
}

// winnersMatch 检查两个获胜者列表是否匹配（顺序无关）
func winnersMatch(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}

	// 排序后比较
	sortedExpected := make([]string, len(expected))
	sortedActual := make([]string, len(actual))
	copy(sortedExpected, expected)
	copy(sortedActual, actual)
	sort.Strings(sortedExpected)
	sort.Strings(sortedActual)

	for i := range sortedExpected {
		if sortedExpected[i] != sortedActual[i] {
			return false
		}
	}
	return true
}
