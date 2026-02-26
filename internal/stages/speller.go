package stages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func spellerTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "speller",
		Timeout:       60 * time.Second,
		TestFunc:      testSpeller,
		RequiredFiles: []string{"dictionary.c"},
		CompileStep: &tester_definition.CompileStep{
			Language: "make",
			Output:   "speller",
		},
	}
}

func testSpeller(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 使用 CS50 提供的测试目录进行测试
	// 每个测试目录包含 dict 和 text 文件
	testCases := []struct {
		name     string
		dir      string   // 测试目录
		expected []string // 期望的拼错单词
		notIn    []string // 不应该出现的单词
	}{
		{
			name:     "handles basic words properly",
			dir:      "basic",
			expected: []string{}, // 所有单词都在字典中
		},
		{
			name:     "handles min length (1-char) words",
			dir:      "min_length",
			expected: []string{},
		},
		{
			name:     "handles max length (45-char) words",
			dir:      "max_length",
			expected: []string{},
		},
		{
			name:     "spell-checks case-insensitively",
			dir:      "case",
			expected: []string{}, // 所有大小写变体都应该匹配
		},
		{
			name:     "handles substrings properly",
			dir:      "substring",
			expected: []string{"ca", "cats", "caterpill", "caterpillars"},
			notIn:    []string{"cat", "caterpillar"},
		},
	}

	for _, tc := range testCases {
		logger.Infof("Testing %s...", tc.name)

		dictPath := filepath.Join(tc.dir, "dict")
		textPath := filepath.Join(tc.dir, "text")

		// 检查测试目录是否存在
		if !harness.FileExists(dictPath) || !harness.FileExists(textPath) {
			return fmt.Errorf("test directory %s not found (missing dict or text)", tc.dir)
		}

		// 运行 speller
		cmd := exec.Command("./speller", dictPath, textPath)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("speller failed on %s: %s\n%s", tc.dir, err, string(out))
		}

		output := string(out)

		// 提取拼错的单词
		misspelled := extractMisspelledWords(output)

		// 检查期望的拼错单词
		for _, word := range tc.expected {
			if !contains(misspelled, word) {
				return fmt.Errorf("expected '%s' to be marked as misspelled", word)
			}
		}

		// 检查不应该出现的单词
		for _, word := range tc.notIn {
			if contains(misspelled, word) {
				return fmt.Errorf("'%s' should not be marked as misspelled", word)
			}
		}

		// 如果期望为空，确保没有拼错的单词
		if len(tc.expected) == 0 && len(misspelled) > 0 {
			return fmt.Errorf("expected no misspelled words, but got: %v", misspelled)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 测试撇号处理 - apostrophe 目录有特殊结构
	logger.Infof("Testing handles apostrophes properly...")

	// 测试 with apostrophe in dict, with apostrophe in text
	cmd := exec.Command("./speller", "apostrophe/with/dict", "apostrophe/with/text")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("speller failed on apostrophe/with: %s\n%s", err, string(out))
	}
	misspelled := extractMisspelledWords(string(out))
	if len(misspelled) > 0 {
		return fmt.Errorf("apostrophe test failed: expected no misspelled words, got: %v", misspelled)
	}
	logger.Successf("✓ handles apostrophes properly")

	// 测试大字典 (可选，验证性能)
	logger.Infof("Testing handles large dictionary...")
	if harness.FileExists("large/dict") && harness.FileExists("large/text") {
		cmd := exec.Command("./speller", "large/dict", "large/text")
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("speller failed on large dictionary: %s\n%s", err, string(out))
		}
		// 只检查程序能正常运行完成，不检查具体输出
		logger.Successf("✓ handles large dictionary")
	}

	// 内存检查 (valgrind) - 如果可用
	logger.Infof("Testing program is free of memory errors...")
	if _, err := exec.LookPath("valgrind"); err != nil {
		logger.Infof("valgrind not available, skipping memory check")
	} else {
		// 使用 basic 目录进行内存检查
		cmd := exec.Command("valgrind", "--error-exitcode=1", "--leak-check=full",
			"--show-leak-kinds=all", "--errors-for-leak-kinds=all", "-q",
			"./speller", "basic/dict", "basic/text")
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("program has memory errors:\n%s", string(out))
		}
		logger.Successf("✓ program is free of memory errors")
	}

	// 清理编译产物
	cleanCmd := exec.Command("make", "clean")
	cleanCmd.Dir = workDir
	cleanCmd.Run()
	os.Remove(filepath.Join(workDir, "speller"))
	os.Remove(filepath.Join(workDir, "speller.o"))
	os.Remove(filepath.Join(workDir, "dictionary.o"))

	logger.Successf("All speller tests passed!")
	return nil
}

// extractMisspelledWords 从 speller 输出中提取拼错的单词
func extractMisspelledWords(output string) []string {
	lines := strings.Split(output, "\n")
	var misspelled []string
	inMisspelled := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "MISSPELLED WORDS" {
			inMisspelled = true
			continue
		}
		if strings.HasPrefix(line, "WORDS MISSPELLED:") {
			break
		}
		if inMisspelled && line != "" {
			misspelled = append(misspelled, line)
		}
	}

	return misspelled
}
