package stages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func marioLessTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "mario-less",
		Timeout:       30 * time.Second,
		TestFunc:      testMarioLess,
		RequiredFiles: []string{"mario.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "mario.c",
			Output:           "mario",
			IncludeParentDir: true,
		},
	}
}

func testMarioLess(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试拒绝无效输入 (对齐 CS50 check50)
	// 使用交互模式: Start() -> SendLine() -> Reject()
	rejectTests := []struct {
		input string
		name  string
	}{
		{"-1", "rejects a height of -1"},
		{"0", "rejects a height of 0"},
		{"foo", "rejects a non-numeric height of \"foo\""},
		{"", "rejects a non-numeric height of \"\""},
	}

	for _, tc := range rejectTests {
		logger.Infof("Testing %s...", tc.name)

		// 使用交互模式: 启动程序 -> 发送输入 -> 检查拒绝（程序还在运行）
		r := runner.Run(workDir, "mario").
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

	// 4. 测试有效输入（使用 txt 文件作为期望输出，对齐 CS50）
	validTests := []struct {
		height  string
		txtFile string
		name    string
	}{
		{"1", "1.txt", "handles a height of 1 correctly"},
		{"2", "2.txt", "handles a height of 2 correctly"},
		{"8", "8.txt", "handles a height of 8 correctly"},
	}

	for _, tc := range validTests {
		logger.Infof("Testing %s...", tc.name)

		// 读取期望输出文件
		expectedBytes, err := os.ReadFile(filepath.Join(workDir, tc.txtFile))
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", tc.txtFile, err)
		}
		expected := strings.TrimSpace(string(expectedBytes))

		r := runner.Run(workDir, "mario").
			WithTimeout(5 * time.Second).
			Stdin(tc.height).
			Stdout(expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 5. 测试拒绝后接受 (CS50 特有测试)
	logger.Infof("Testing rejects -1, then accepts 2...")
	expectedBytes, err := os.ReadFile(filepath.Join(workDir, "2.txt"))
	if err != nil {
		return fmt.Errorf("failed to read 2.txt: %v", err)
	}
	expected := strings.TrimSpace(string(expectedBytes))

	r := runner.Run(workDir, "mario").
		WithTimeout(5 * time.Second).
		Stdin("-1\n2\n").
		Stdout(expected).
		Exit(0)

	if err := r.Error(); err != nil {
		return fmt.Errorf("rejects -1 then accepts 2: %v", err)
	}
	logger.Successf("✓ rejects -1, then accepts 2")

	logger.Successf("All mario-less tests passed!")
	return nil
}
