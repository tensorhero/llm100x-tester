package stages

import (
	"fmt"
	"strings"
	"time"

	"github.com/hellobyte-dev/tester-utils/random"
	"github.com/hellobyte-dev/tester-utils/runner"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

// Scrabble 字母分值表（对齐 CS50）
var POINTS = []int{1, 3, 3, 2, 1, 4, 2, 4, 1, 8, 5, 1, 3, 1, 1, 3, 10, 1, 1, 1, 1, 4, 4, 8, 4, 10}

// 生成所有分值为 1 的字母
func getOnePointLetters() []string {
	letters := []string{}
	for i := 0; i < 26; i++ {
		if POINTS[i] == 1 {
			letters = append(letters, string(rune('a'+i)))
		}
	}
	return letters
}

func scrabbleTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "scrabble",
		Timeout:       30 * time.Second,
		TestFunc:      testScrabble,
		RequiredFiles: []string{"scrabble.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "scrabble.c",
			Output:           "scrabble",
			IncludeParentDir: true,
		},
	}
}

func testScrabble(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例（完全对齐 CS50 check50）
	testCases := []struct {
		word1    string
		word2    string
		expected string
		name     string
	}{
		// CS50 check50: tie_letter_case
		{
			"LETTERCASE", "lettercase",
			"Tie!",
			"handles letter cases correctly",
		},
		// CS50 check50: tie_punctuation
		{
			"Punctuation!?!?", "punctuation",
			"Tie!",
			"handles punctuation correctly",
		},
		// CS50 check50: test1
		{
			"Question?", "Question!",
			"Tie!",
			"correctly identifies 'Question?' and 'Question!' as a tie",
		},
		// CS50 check50: test2
		{
			"drawing", "illustration",
			"Tie!",
			"correctly identifies 'drawing' and 'illustration' as a tie",
		},
		// CS50 check50: test3
		{
			"Oh,", "hai!",
			"Player 2 wins!",
			"correctly identifies 'hai!' as winner over 'Oh,'",
		},
		// CS50 check50: test4
		{
			"COMPUTER", "science",
			"Player 1 wins!",
			"correctly identifies 'COMPUTER' as winner over 'science'",
		},
		// CS50 check50: test5
		{
			"Scrabble", "wiNNeR",
			"Player 1 wins!",
			"correctly identifies 'Scrabble' as winner over 'wiNNeR'",
		},
		// CS50 check50: test6
		{
			"pig", "dog",
			"Player 1 wins!",
			"correctly identifies 'pig' as winner over 'dog'",
		},
		// CS50 check50: complex_case
		{
			"figure?", "Skating!",
			"Player 2 wins!",
			"correctly identifies 'Skating!' as winner over 'figure?'",
		},
	}

	for _, tc := range testCases {
		logger.Infof("Testing %s...", tc.name)

		// 发送两行输入：word1 + word2
		input := fmt.Sprintf("%s\n%s\n", tc.word1, tc.word2)

		r := runner.Run(workDir, "scrabble").
			WithTimeout(5 * time.Second).
			Stdin(input).
			Stdout(tc.expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("%s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 4. CS50 check50: test_strict_order() - 随机字母顺序测试
	// 测试相邻字母的分数比较（例如 'a' vs 'b', 'c' vs 'd'）
	logger.Infof("Testing random letter pairs (test_strict_order)...")

	// 随机选择5对相邻字母进行测试
	numTests := 5
	if len(POINTS)-1 < numTests {
		numTests = len(POINTS) - 1
	}

	// 使用 random.RandomInts 选择不重复的索引
	indices := random.RandomInts(0, len(POINTS)-1, numTests)

	for _, i := range indices {
		letter1 := string(rune('a' + i))
		letter2 := string(rune('a' + i + 1))

		// 计算预期结果
		var expected string
		pointsDiff := POINTS[i+1] - POINTS[i]
		if pointsDiff > 0 {
			expected = "Player 2 wins!"
		} else if pointsDiff < 0 {
			expected = "Player 1 wins!"
		} else {
			expected = "Tie!"
		}

		input := fmt.Sprintf("%s\n%s\n", letter1, letter2)

		r := runner.Run(workDir, "scrabble").
			WithTimeout(5 * time.Second).
			Stdin(input).
			Stdout(expected).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test_strict_order failed for '%s' vs '%s': %v", letter1, letter2, err)
		}

		logger.Debugf("✓ '%s' vs '%s' → %s", letter1, letter2, expected)
	}
	logger.Successf("✓ random letter pairs test passed")

	// 5. CS50 check50: test_scoring_accuracy() - 精确计分测试
	// 验证单个字母的分数计算是否准确
	logger.Infof("Testing scoring accuracy (test_scoring_accuracy)...")

	onePointLetters := getOnePointLetters()

	// 随机选择5个字母进行计分验证
	numScoreTests := 5
	if len(POINTS) < numScoreTests {
		numScoreTests = len(POINTS)
	}

	letterIndices := random.RandomInts(0, 26, numScoreTests)

	for _, i := range letterIndices {
		letter := string(rune('a' + i))
		points := POINTS[i]

		// 创建一个由多个1分字母组成的单词，总分等于测试字母的分数
		// 例如：如果 'b' = 3分，就用 "aaa"（3个1分字母）来对比
		if len(onePointLetters) == 0 {
			continue // 如果没有1分字母，跳过此测试
		}

		onePointLetter := onePointLetters[random.RandomInt(0, len(onePointLetters))]
		word := strings.Repeat(onePointLetter, points)

		input := fmt.Sprintf("%s\n%s\n", letter, word)

		r := runner.Run(workDir, "scrabble").
			WithTimeout(5 * time.Second).
			Stdin(input).
			Stdout("Tie!").
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test_scoring_accuracy failed for '%s' (points=%d) vs '%s': %v",
				letter, points, word, err)
		}

		logger.Debugf("✓ '%s' (%d points) vs '%s' (%dx%d) → Tie",
			letter, points, word, points, 1)
	}
	logger.Successf("✓ scoring accuracy test passed")

	logger.Successf("All scrabble tests passed!")
	return nil
}
