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

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func runoffTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "runoff",
		Timeout:       60 * time.Second,
		TestFunc:      testRunoff,
		RequiredFiles: []string{"runoff.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "runoff.c",
			Output:           "runoff",
			IncludeParentDir: true,
		},
	}
}

func testRunoff(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 构建测试套件
	// 读取学生的 runoff.c，将 main 重命名为 distro_main
	runoffCode, err := harness.ReadFile("runoff.c")
	if err != nil {
		return fmt.Errorf("could not read runoff.c: %v", err)
	}

	// 使用正则替换 main 函数
	mainRegex := regexp.MustCompile(`int\s+main\s*\(`)
	modifiedCode := mainRegex.ReplaceAllString(string(runoffCode), "int distro_main(")

	// 从学生目录读取测试代码
	testCodeBytes, err := harness.ReadFile("runoff_test.c")
	if err != nil {
		return fmt.Errorf("runoff_test.c does not exist: %v", err)
	}
	testCode := string(testCodeBytes)
	combinedCode := modifiedCode + "\n" + testCode
	testFilePath := filepath.Join(workDir, "runoff_combined_test.c")
	if err := os.WriteFile(testFilePath, []byte(combinedCode), 0644); err != nil {
		return fmt.Errorf("could not write test file: %v", err)
	}

	// 编译测试程序
	logger.Infof("Compiling test harness...")
	cmd := exec.Command("clang", "-o", "runoff_test", "runoff_combined_test.c", "-I..", "-lm", "-Wall")
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
		{"0", "0", "true", "vote returns true when given name of valid candidate"},
		{"0", "1", "false", "vote returns false when given name of invalid candidate"},
		{"0", "2", "2", "vote correctly sets first preference"},
		{"0", "3", "0", "vote correctly sets third preference"},
		{"0", "4", "1 0 2", "vote correctly sets all preferences"},
	}

	for _, tc := range voteTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 5. 运行 tabulate 函数测试
	tabulateTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"1", "5", "3 3 1 0 ", "tabulate counts votes when all candidates are in election"},
		{"1", "6", "3 3 1 0 ", "tabulate counts votes when one candidate is eliminated"},
		{"1", "7", "3 4 0 0 ", "tabulate counts votes when multiple candidates are eliminated"},
		{"1", "22", "3 4 0 0 ", "tabulate counts votes when multiple rounds have occurred"},
	}

	for _, tc := range tabulateTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 6. 运行 print_winner 函数测试
	logger.Infof("Testing print_winner prints name of candidate with > 50%% votes...")
	r := runner.Run(workDir, "runoff_test", "2", "8").
		WithTimeout(5 * time.Second).
		Execute().
		Exit(0)
	if err := r.Error(); err != nil {
		return fmt.Errorf("test failed for print_winner: %v", err)
	}
	stdout := strings.TrimSpace(r.GetStdout())
	if stdout != "Bob" {
		return fmt.Errorf("print_winner did not print correct winner: expected 'Bob', got '%s'", stdout)
	}
	logger.Successf("✓ print_winner prints name of candidate with > 50%% votes")

	printWinnerTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"2", "9", "Bob\ntrue", "print_winner returns true when someone has > 50% votes"},
		{"2", "10", "false", "print_winner returns false when no one has > 50% votes"},
		{"2", "11", "false", "print_winner returns false when exactly 50% votes"},
	}

	for _, tc := range printWinnerTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 7. 运行 find_min 函数测试
	findMinTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"2", "12", "1", "find_min returns minimum votes"},
		{"2", "13", "7", "find_min returns minimum when all tied"},
		{"2", "14", "4", "find_min ignores eliminated candidates"},
	}

	for _, tc := range findMinTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 8. 运行 is_tie 函数测试
	isTieTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"2", "15", "true", "is_tie returns true when all candidates are tied"},
		{"2", "16", "false", "is_tie returns false when not tied"},
		{"2", "17", "false", "is_tie returns false when only some candidates are tied"},
		{"2", "18", "true", "is_tie ignores eliminated candidates when checking tie"},
	}

	for _, tc := range isTieTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 9. 运行 eliminate 函数测试
	eliminateTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"2", "19", "false false false true ", "eliminate eliminates candidate with minimum votes"},
		{"2", "20", "true false true false ", "eliminate eliminates multiple candidates tied for last"},
		{"2", "21", "true false true false ", "eliminate correctly identifies who to eliminate after some already eliminated"},
	}

	for _, tc := range eliminateTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "runoff_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 清理测试文件
	os.Remove(testFilePath)
	os.Remove(filepath.Join(workDir, "runoff_test"))

	logger.Successf("All runoff tests passed!")
	return nil
}

// parseRunoffWinners 从输出中解析获胜者名单
func parseRunoffWinners(output string) []string {
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

// runoffWinnersMatch 检查两个获胜者列表是否匹配（顺序无关）
func runoffWinnersMatch(expected, actual []string) bool {
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
