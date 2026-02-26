package stages

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

// 期望的哈希值 (CS50 提供的)
// permits either truncating or rounding the floats to ints
var hashesHalf = []string{
	"268f2deee976fcc8dcc915be63cd3d4ac003a73a4b8bfd6f7b95c441a42ed1ec",
	"9a92829dcd2de343235607de1db01d374a44bd58738cb4070e354dead02e50c1",
}

var hashesTenth = []string{
	"4481b5a438d359718000dfd58e2a32a7b109eb4a5590e0650c6bd295979c64fc",
	"dae24291174811b2df95f6836e023e855c77bdc5bee3294542ba4bf1b95de2cc",
}

var hashesDouble = []string{
	"3d83603745302935c067379b704573e5addb4356ad407041f0a698070e6e4e7b",
}

func volumeTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "volume",
		Timeout:       30 * time.Second,
		TestFunc:      testVolume,
		RequiredFiles: []string{"volume.c", "input.wav"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "volume.c",
			Output:           "volume",
			IncludeParentDir: true,
		},
	}
}

func testVolume(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试 factor 0.5
	logger.Infof("Testing reduces audio volume, factor of 0.5 correctly...")
	outputPath := filepath.Join(workDir, "output.wav")
	cmd := exec.Command("./volume", "input.wav", "output.wav", "0.5")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("volume failed with factor 0.5: %s\n%s", err, string(out))
	}

	hash, err := hashFile(outputPath)
	if err != nil {
		return fmt.Errorf("could not hash output.wav: %v", err)
	}

	if !contains(hashesHalf, hash) {
		return fmt.Errorf("audio is not correctly altered, factor of 0.5 (hash: %s)", hash)
	}
	logger.Successf("✓ reduces audio volume, factor of 0.5 correctly")

	// 5. 测试 factor 0.1
	logger.Infof("Testing reduces audio volume, factor of 0.1 correctly...")
	cmd = exec.Command("./volume", "input.wav", "output.wav", "0.1")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("volume failed with factor 0.1: %s\n%s", err, string(out))
	}

	hash, err = hashFile(outputPath)
	if err != nil {
		return fmt.Errorf("could not hash output.wav: %v", err)
	}

	if !contains(hashesTenth, hash) {
		return fmt.Errorf("audio is not correctly altered, factor of 0.1 (hash: %s)", hash)
	}
	logger.Successf("✓ reduces audio volume, factor of 0.1 correctly")

	// 6. 测试 factor 2
	logger.Infof("Testing increases audio volume, factor of 2 correctly...")
	cmd = exec.Command("./volume", "input.wav", "output.wav", "2")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("volume failed with factor 2: %s\n%s", err, string(out))
	}

	hash, err = hashFile(outputPath)
	if err != nil {
		return fmt.Errorf("could not hash output.wav: %v", err)
	}

	if !contains(hashesDouble, hash) {
		return fmt.Errorf("audio is not correctly altered, factor of 2 (hash: %s)", hash)
	}
	logger.Successf("✓ increases audio volume, factor of 2 correctly")

	// 清理编译产物
	os.Remove(outputPath)
	os.Remove(filepath.Join(workDir, "volume"))

	logger.Successf("All volume tests passed!")
	return nil
}

// hashFile 计算文件的 SHA256 哈希
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// contains 检查切片是否包含指定字符串
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
