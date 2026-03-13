package stages

import (
	"fmt"
	"time"

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func readabilityTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "readability",
		Timeout:       30 * time.Second,
		TestFunc:      testReadability,
		RequiredFiles: []string{"readability.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "readability.c",
			Output:           "readability",
			IncludeParentDir: true,
		},
	}
}

func testReadability(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 测试用例（完全对齐 CS50 check50）
	testCases := []struct {
		input         string
		expectedGrade string
		name          string
	}{
		{
			"In my younger and more vulnerable years my father gave me some advice that I've been turning over in my mind ever since.",
			"Grade 7",
			"handles single sentence with multiple words",
		},
		{
			"There are more things in Heaven and Earth, Horatio, than are dreamt of in your philosophy.",
			"Grade 9",
			"handles punctuation within a single sentence",
		},
		{
			`Alice was beginning to get very tired of sitting by her sister on the bank, and of having nothing to do: once or twice she had peeped into the book her sister was reading, but it had no pictures or conversations in it, "and what is the use of a book," thought Alice "without pictures or conversation?"`,
			"Grade 8",
			"handles more complex single sentence",
		},
		{
			"Harry Potter was a highly unusual boy in many ways. For one thing, he hated the summer holidays more than any other time of year. For another, he really wanted to do his homework, but was forced to do it in secret, in the dead of the night. And he also happened to be a wizard.",
			"Grade 5",
			"handles multiple sentences",
		},
		{
			"It was a bright cold day in April, and the clocks were striking thirteen. Winston Smith, his chin nuzzled into his breast in an effort to escape the vile wind, slipped quickly through the glass doors of Victory Mansions, though not quickly enough to prevent a swirl of gritty dust from entering along with him.",
			"Grade 10",
			"handles multiple more complex sentences",
		},
		{
			"When he was nearly thirteen, my brother Jem got his arm badly broken at the elbow. When it healed, and Jem's fears of never being able to play football were assuaged, he was seldom self-conscious about his injury. His left arm was somewhat shorter than his right; when he stood or walked, the back of his hand was at right angles to his body, his thumb parallel to his thigh.",
			"Grade 8",
			"handles longer passages",
		},
		{
			"Congratulations! Today is your day. You're off to Great Places! You're off and away!",
			"Grade 3",
			"handles multiple sentences with different punctuation",
		},
		{
			"Would you like them here or there? I would not like them here or there. I would not like them anywhere.",
			"Grade 2",
			"handles questions in passage",
		},
		{
			"One fish. Two fish. Red fish. Blue fish.",
			"Before Grade 1",
			"handles reading level before Grade 1",
		},
		{
			"A large class of computational problems involve the determination of properties of graphs, digraphs, integers, arrays of integers, finite families of finite sets, boolean formulas and elements of other countable domains.",
			"Grade 16+",
			"handles reading level at Grade 16+",
		},
	}

	for _, tc := range testCases {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "readability").
			WithTimeout(5 * time.Second).
			Stdin(tc.input).
			Stdout(tc.expectedGrade).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	logger.Successf("All readability tests passed!")
	return nil
}
