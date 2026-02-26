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

func filterLessTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "filter-less",
		Timeout:       60 * time.Second,
		TestFunc:      testFilterLess,
		RequiredFiles: []string{"helpers.c", "bmp.h", "helpers.h", "testing.c"},
	}
}

func testFilterLess(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 编译 filter
	logger.Infof("Compiling filter...")
	cmd := exec.Command("clang",
		"-ggdb3", "-gdwarf-4", "-O0", "-Qunused-arguments",
		"-std=c11", "-Wall", "-Werror", "-Wextra",
		"-Wno-sign-compare", "-Wno-unused-parameter", "-Wno-unused-variable",
		"-lm", "-o", "testing", "testing.c", "helpers.c")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("filter does not compile: %s\n%s", err, string(out))
	}
	logger.Successf("filter compiles")

	// 5. 运行所有 grayscale 测试
	grayscaleTests := []struct {
		test     int
		name     string
		expected string
	}{
		{0, "grayscale correctly filters single pixel with whole number average", "50 50 50\n"},
		{1, "grayscale correctly filters single pixel without whole number average", "28 28 28\n"},
		{2, "grayscale leaves alone pixels that are already gray", "50 50 50\n"},
		{3, "grayscale correctly filters simple 3x3 image", strings.Repeat("85 85 85\n", 9)},
		{4, "grayscale correctly filters more complex 3x3 image",
			"20 20 20\n50 50 50\n80 80 80\n" +
				"127 127 127\n137 137 137\n147 147 147\n" +
				"210 210 210\n230 230 230\n248 248 248\n"},
		{5, "grayscale correctly filters 4x4 image",
			"20 20 20\n50 50 50\n80 80 80\n110 110 110\n" +
				"127 127 127\n137 137 137\n147 147 147\n157 157 157\n" +
				"204 204 204\n214 214 214\n234 234 234\n251 251 251\n" +
				"56 56 56\n0 0 0\n255 255 255\n85 85 85\n"},
	}

	for _, tc := range grayscaleTests {
		logger.Infof("Testing %s...", tc.name)
		if err := runFilterTest(workDir, 0, tc.test, tc.expected); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}
		logger.Successf("✓ %s", tc.name)
	}

	// 6. 运行所有 sepia 测试
	sepiaTests := []struct {
		test     int
		name     string
		expected string
	}{
		{0, "sepia correctly filters single pixel", "56 50 39\n"},
		{3, "sepia correctly filters simple 3x3 image",
			"100 89 69\n100 89 69\n100 89 69\n" +
				"196 175 136\n196 175 136\n196 175 136\n" +
				"48 43 33\n48 43 33\n48 43 33\n"},
		{4, "sepia correctly filters more complex 3x3 image",
			"25 22 17\n66 58 45\n106 94 74\n" +
				"170 151 118\n183 163 127\n197 175 136\n" +
				"255 251 195\n255 255 214\n255 255 232\n"},
		{5, "sepia correctly filters 4x4 image",
			"25 22 17\n66 58 45\n106 94 74\n147 131 102\n" +
				"170 151 118\n183 163 127\n197 175 136\n210 187 146\n" +
				"255 244 190\n255 255 199\n255 255 218\n255 255 235\n" +
				"58 52 40\n0 0 0\n255 255 239\n115 102 80\n"},
	}

	for _, tc := range sepiaTests {
		logger.Infof("Testing %s...", tc.name)
		if err := runFilterTest(workDir, 1, tc.test, tc.expected); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}
		logger.Successf("✓ %s", tc.name)
	}

	// 7. 运行所有 reflect 测试
	reflectTests := []struct {
		test     int
		name     string
		expected string
	}{
		{0, "reflect correctly filters 1x2 image", "0 0 255\n255 0 0\n"},
		{1, "reflect correctly filters 1x3 image", "0 0 255\n0 255 0\n255 0 0\n"},
		{2, "reflect correctly filters image that is its own mirror image",
			"255 0 0\n255 0 0\n255 0 0\n" +
				"0 255 0\n0 255 0\n0 255 0\n" +
				"0 0 255\n0 0 255\n0 0 255\n"},
		{3, "reflect correctly filters 3x3 image",
			"70 80 90\n40 50 60\n10 20 30\n" +
				"130 150 160\n120 140 150\n110 130 140\n" +
				"240 250 255\n220 230 240\n200 210 220\n"},
		{4, "reflect correctly filters 4x4 image",
			"100 110 120\n70 80 90\n40 50 60\n10 20 30\n" +
				"140 160 170\n130 150 160\n120 140 150\n110 130 140\n" +
				"245 254 253\n225 234 243\n205 214 223\n195 204 213\n" +
				"85 85 85\n255 255 255\n0 0 0\n50 28 90\n"},
	}

	for _, tc := range reflectTests {
		logger.Infof("Testing %s...", tc.name)
		if err := runFilterTest(workDir, 2, tc.test, tc.expected); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}
		logger.Successf("✓ %s", tc.name)
	}

	// 8. 运行所有 blur 测试
	blurTests := []struct {
		test     int
		name     string
		expected string
	}{
		{0, "blur correctly filters middle pixel", "127 140 149\n"},
		{1, "blur correctly filters pixel on edge", "80 95 105\n"},
		{2, "blur correctly filters pixel in corner", "70 85 95\n"},
		{3, "blur correctly filters 3x3 image",
			"70 85 95\n80 95 105\n90 105 115\n" +
				"117 130 140\n127 140 149\n137 150 159\n" +
				"163 178 188\n170 185 194\n178 193 201\n"},
		{4, "blur correctly filters 4x4 image",
			"70 85 95\n80 95 105\n100 115 125\n110 125 135\n" +
				"113 126 136\n123 136 145\n142 155 163\n152 165 173\n" +
				"113 119 136\n143 151 164\n156 166 171\n180 190 194\n" +
				"113 112 132\n155 156 171\n169 174 177\n203 207 209\n"},
	}

	for _, tc := range blurTests {
		logger.Infof("Testing %s...", tc.name)
		if err := runFilterTest(workDir, 3, tc.test, tc.expected); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}
		logger.Successf("✓ %s", tc.name)
	}

	// 清理编译产物
	os.Remove(filepath.Join(workDir, "testing"))

	logger.Successf("All filter tests passed!")
	return nil
}

func runFilterTest(workDir string, function, test int, expected string) error {
	cmd := exec.Command("./testing", fmt.Sprintf("%d", function), fmt.Sprintf("%d", test))
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("test failed: %s\n%s", err, string(out))
	}

	actual := string(out)
	if actual != expected {
		return fmt.Errorf("output mismatch\nExpected:\n%s\nGot:\n%s", expected, actual)
	}

	return nil
}
