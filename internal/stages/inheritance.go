package stages

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func inheritanceTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "inheritance",
		Timeout:       30 * time.Second,
		TestFunc:      testInheritance,
		RequiredFiles: []string{"inheritance.c"},
		CompileStep: &tester_definition.CompileStep{
			Language: "c",
			Source:   "inheritance.c",
			Output:   "inheritance",
			Flags:    []string{"-ggdb3", "-gdwarf-4", "-O0", "-Qunused-arguments", "-std=c11", "-Wextra", "-Wno-sign-compare", "-Wno-unused-parameter", "-Wno-unused-variable"},
		},
	}
}

func testInheritance(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 创建测试程序
	// 读取学生的 inheritance.c，将 main 重命名为 distro_main
	inheritanceCode, err := harness.ReadFile("inheritance.c")
	if err != nil {
		return fmt.Errorf("could not read inheritance.c: %v", err)
	}

	// 使用正则替换 main 函数
	mainRegex := regexp.MustCompile(`int\s+main\s*\(`)
	modifiedCode := mainRegex.ReplaceAllString(string(inheritanceCode), "int distro_main(")

	// 读取测试代码
	testCodeBytes, err := harness.ReadFile("inheritance_test.c")
	if err != nil {
		return fmt.Errorf("inheritance_test.c does not exist: %v", err)
	}
	testCode := string(testCodeBytes)

	// 写入组合的测试文件
	combinedCode := modifiedCode + "\n" + testCode
	testFilePath := filepath.Join(workDir, "inheritance_combined_test.c")
	if err := os.WriteFile(testFilePath, []byte(combinedCode), 0644); err != nil {
		return fmt.Errorf("could not write test file: %v", err)
	}

	// 编译测试程序
	logger.Infof("Compiling test harness...")
	cmd := exec.Command("clang",
		"-ggdb3", "-gdwarf-4", "-O0", "-Qunused-arguments",
		"-std=c11", "-Wall", "-Wextra",
		"-Wno-sign-compare", "-Wno-unused-parameter", "-Wno-unused-variable",
		"-lm", "-o", "inheritance_test", "inheritance_combined_test.c")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("test harness does not compile: %s\n%s", err, string(out))
	}
	logger.Successf("test harness compiles")

	// 4. 测试正确的家族大小
	logger.Infof("Testing correct family size...")
	cmd = exec.Command("./inheritance_test")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("test program failed: %s\n%s", err, string(out))
	}

	output := strings.TrimSpace(string(out))
	if !strings.Contains(output, "size_true") {
		return fmt.Errorf("incorrect family size: expected 3 generations")
	}
	logger.Successf("✓ correct family size")

	// 5. 测试等位基因正确继承
	logger.Infof("Testing alleles inherited correctly...")
	if !strings.Contains(output, "allele_true") {
		return fmt.Errorf("alleles not inherited correctly from parents")
	}
	logger.Successf("✓ alleles inherited correctly")

	// 6. 多次运行验证一致性
	logger.Infof("Testing multiple runs for consistency...")
	for i := 0; i < 5; i++ {
		cmd = exec.Command("./inheritance_test")
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run %d failed: %s\n%s", i+1, err, string(out))
		}
		output := strings.TrimSpace(string(out))
		if !strings.Contains(output, "size_true") || !strings.Contains(output, "allele_true") {
			return fmt.Errorf("run %d: got %s", i+1, output)
		}
	}
	logger.Successf("✓ multiple runs consistent")

	// 7. 内存检查 (valgrind) - 如果可用
	logger.Infof("Testing program is free of memory errors...")
	if _, err := exec.LookPath("valgrind"); err != nil {
		logger.Infof("valgrind not available, skipping memory check")
	} else {
		cmd = exec.Command("valgrind", "--error-exitcode=1", "--leak-check=full",
			"--show-leak-kinds=all", "--errors-for-leak-kinds=all", "-q", "./inheritance_test")
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("program has memory errors:\n%s", string(out))
		}
		logger.Successf("✓ program is free of memory errors")
	}

	// 清理编译产物
	os.Remove(filepath.Join(workDir, "inheritance"))
	os.Remove(filepath.Join(workDir, "inheritance_test"))
	os.Remove(filepath.Join(workDir, "inheritance_combined_test.c"))

	logger.Successf("All inheritance tests passed!")
	return nil
}
