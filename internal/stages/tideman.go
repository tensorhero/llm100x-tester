package stages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func tidemanTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "tideman",
		Timeout:       60 * time.Second,
		TestFunc:      testTideman,
		RequiredFiles: []string{"tideman.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "tideman.c",
			Output:           "tideman",
			IncludeParentDir: true,
		},
	}
}

func testTideman(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 构建测试套件
	// 读取学生的 tideman.c，将 main 重命名为 distro_main
	tidemanCode, err := harness.ReadFile("tideman.c")
	if err != nil {
		return fmt.Errorf("could not read tideman.c: %v", err)
	}

	// 使用正则替换 main 函数
	mainRegex := regexp.MustCompile(`int\s+main\s*\(`)
	modifiedCode := mainRegex.ReplaceAllString(string(tidemanCode), "int distro_main(")

	// 从学生目录读取测试代码
	testCodeBytes, err := harness.ReadFile("tideman_test.c")
	if err != nil {
		return fmt.Errorf("tideman_test.c does not exist: %v", err)
	}
	testCode := string(testCodeBytes)
	combinedCode := modifiedCode + "\n" + testCode
	testFilePath := filepath.Join(workDir, "tideman_combined_test.c")
	if err := os.WriteFile(testFilePath, []byte(combinedCode), 0644); err != nil {
		return fmt.Errorf("could not write test file: %v", err)
	}

	// 编译测试程序
	logger.Infof("Compiling test harness...")
	cmd := exec.Command("clang", "-o", "tideman_test", "tideman_combined_test.c", "-I..", "-lm", "-Wall")
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
		{"0", "0", "true", "vote returns true when given name of candidate"},
		{"0", "1", "false", "vote returns false when given name of invalid candidate"},
		{"0", "2", "1", "vote correctly sets rank for first preference"},
		{"0", "2", "1 2 0", "vote correctly sets rank for all preferences"},
	}

	for _, tc := range voteTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "tideman_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 5. 运行 record_preferences 函数测试
	recordPrefsTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"0", "3", "0 0 0 1 0 1 1 0 0 ", "record_preferences correctly sets preferences for first voter"},
		{"0", "4", "0 2 2 4 0 5 3 5 0", "record_preferences correctly sets preferences for all voters"},
	}

	for _, tc := range recordPrefsTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "tideman_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 6. 运行 add_pairs 函数测试
	addPairsTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"1", "5", "3", "add_pairs generates correct pair count when no ties"},
		{"2", "5", "2", "add_pairs generates correct pair count when ties exist"},
		{"1", "6", "true true true ", "add_pairs fills pairs array with winning pairs"},
		{"1", "7", "0", "add_pairs does not fill pairs array with losing pairs"},
	}

	for _, tc := range addPairsTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "tideman_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 7. 运行 sort_pairs 函数测试
	logger.Infof("Testing sort_pairs sorts pairs of candidates by margin of victory...")
	r := runner.Run(workDir, "tideman_test", "3", "8").
		WithTimeout(5 * time.Second).
		Execute().
		Stdout("0 2 0 1 2 1 ").
		Exit(0)
	if err := r.Error(); err != nil {
		return fmt.Errorf("test failed for sort_pairs: %v", err)
	}
	logger.Successf("✓ sort_pairs sorts pairs of candidates by margin of victory")

	// 8. 运行 lock_pairs 函数测试
	lockPairsTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"5", "16", "false false false true false true false false false false false false false false false false false false false true false false true false false ", "lock_pairs locks all pairs when no cycles"},
		{"6", "14", "false true false false false false false false false false true false false false false false false false false false false false false true false false true true false false false false false false false false ", "lock_pairs skips final pair if it creates cycle"},
		{"5", "15", "false false false false false false false false true false true false false false false false false false false false false true true false false ", "lock_pairs skips middle pair if it creates a cycle"},
	}

	for _, tc := range lockPairsTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "tideman_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 9. 运行 print_winner 函数测试
	printWinnerTests := []struct {
		setup    string
		test     string
		expected string
		name     string
	}{
		{"4", "12", "Alice", "print_winner prints winner of election when one candidate wins over all others"},
		{"4", "13", "Charlie", "print_winner prints winner of election when some pairs are tied"},
	}

	for _, tc := range printWinnerTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "tideman_test", tc.setup, tc.test).
			WithTimeout(5 * time.Second).
			Execute().
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		stdout := strings.TrimSpace(r.GetStdout())
		if stdout != tc.expected {
			return fmt.Errorf("test failed for %s: expected '%s', got '%s'", tc.name, tc.expected, stdout)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 清理测试文件
	os.Remove(testFilePath)
	os.Remove(filepath.Join(workDir, "tideman_test"))

	logger.Successf("All tideman tests passed!")
	return nil
}
