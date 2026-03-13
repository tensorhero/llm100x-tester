package stages

import (
	"fmt"
	"time"

	"github.com/tensorhero/tester-utils/runner"
	"github.com/tensorhero/tester-utils/test_case_harness"
	"github.com/tensorhero/tester-utils/tester_definition"
)

func substitutionTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "substitution",
		Timeout:       30 * time.Second,
		TestFunc:      testSubstitution,
		RequiredFiles: []string{"substitution.c"},
		CompileStep: &tester_definition.CompileStep{
			Language:         "c",
			Source:           "substitution.c",
			Output:           "substitution",
			IncludeParentDir: true,
		},
	}
}

func testSubstitution(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 加密测试用例（完全对齐 CS50 check50）
	encryptTests := []struct {
		key        string
		plaintext  string
		ciphertext string
		name       string
	}{
		// encrypt1: encrypts "A" as "Z" using ZYXWVUTSRQPONMLKJIHGFEDCBA as key
		{
			"ZYXWVUTSRQPONMLKJIHGFEDCBA", "A", "Z",
			"encrypts \"A\" as \"Z\" using ZYXWVUTSRQPONMLKJIHGFEDCBA as key",
		},
		// encrypt2: encrypts "a" as "z" using ZYXWVUTSRQPONMLKJIHGFEDCBA as key
		{
			"ZYXWVUTSRQPONMLKJIHGFEDCBA", "a", "z",
			"encrypts \"a\" as \"z\" using ZYXWVUTSRQPONMLKJIHGFEDCBA as key",
		},
		// encrypt3: encrypts "ABC" as "NJQ" using NJQSUYBRXMOPFTHZVAWCGILKED as key
		{
			"NJQSUYBRXMOPFTHZVAWCGILKED", "ABC", "NJQ",
			"encrypts \"ABC\" as \"NJQ\" using NJQSUYBRXMOPFTHZVAWCGILKED as key",
		},
		// encrypt4: encrypts "XyZ" as "KeD" using NJQSUYBRXMOPFTHZVAWCGILKED as key
		{
			"NJQSUYBRXMOPFTHZVAWCGILKED", "XyZ", "KeD",
			"encrypts \"XyZ\" as \"KeD\" using NJQSUYBRXMOPFTHZVAWCGILKED as key",
		},
		// encrypt5: encrypts "This is CS50" as "Cbah ah KH50" using YUKFRNLBAVMWZTEOGXHCIPJSQD as key
		{
			"YUKFRNLBAVMWZTEOGXHCIPJSQD", "This is CS50", "Cbah ah KH50",
			"encrypts \"This is CS50\" as \"Cbah ah KH50\" using YUKFRNLBAVMWZTEOGXHCIPJSQD as key",
		},
		// encrypt6: encrypts "This is CS50" as "Cbah ah KH50" using yukfrnlbavmwzteogxhcipjsqd as key (lowercase)
		{
			"yukfrnlbavmwzteogxhcipjsqd", "This is CS50", "Cbah ah KH50",
			"encrypts \"This is CS50\" as \"Cbah ah KH50\" using yukfrnlbavmwzteogxhcipjsqd as key",
		},
		// encrypt7: encrypts "This is CS50" as "Cbah ah KH50" using YUKFRNLBAVMWZteogxhcipjsqd as key (mixed)
		{
			"YUKFRNLBAVMWZteogxhcipjsqd", "This is CS50", "Cbah ah KH50",
			"encrypts \"This is CS50\" as \"Cbah ah KH50\" using YUKFRNLBAVMWZteogxhcipjsqd as key",
		},
		// encrypt8: encrypts all alphabetic characters using DWUSXNPQKEGCZFJBTLYROHIAVM as key
		{
			"DWUSXNPQKEGCZFJBTLYROHIAVM", "The quick brown fox jumps over the lazy dog", "Rqx tokug wljif nja eozby jhxl rqx cdmv sjp",
			"encrypts all alphabetic characters using DWUSXNPQKEGCZFJBTLYROHIAVM as key",
		},
		// encrypt9: does not encrypt non-alphabetical characters
		{
			"DWUSXNPQKEGCZFJBTLYROHIAVM", "Shh... Don't tell!", "Yqq... Sjf'r rxcc!",
			"does not encrypt non-alphabetical characters using DWUSXNPQKEGCZFJBTLYROHIAVM as key",
		},
	}

	for _, tc := range encryptTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "substitution", tc.key).
			WithTimeout(5 * time.Second).
			Stdin(tc.plaintext).
			Stdout(tc.ciphertext).
			Exit(0)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	// 4. 错误处理测试用例
	errorTests := []struct {
		args []string
		name string
	}{
		// handles lack of key
		{
			[]string{},
			"handles lack of key",
		},
		// handles too many arguments
		{
			[]string{"abcdefghijklmnopqrstuvwxyz", "abc"},
			"handles too many arguments",
		},
		// handles invalid key length
		{
			[]string{"QTXDGMKIPV"},
			"handles invalid key length",
		},
		// handles invalid characters in key
		{
			[]string{"ZWGKPMJ^YISHFEXQON[DLUACVT"},
			"handles invalid characters in key",
		},
		// handles duplicate characters in uppercase key
		{
			[]string{"FAZRDTMGQEJPWAXUSKVIYCLONH"},
			"handles duplicate characters in uppercase key",
		},
		// handles duplicate characters in lowercase key
		{
			[]string{"fazrdtmgqejpwaxuskviyclonh"},
			"handles duplicate characters in lowercase key",
		},
		// handles multiple duplicate characters in key
		{
			[]string{"MMCcEFGHIJKLMNOPqRqTUVWXeZ"},
			"handles multiple duplicate characters in key",
		},
		// handles a single mixed-case duplicate character in key
		{
			[]string{"ABCDEFGHIJKLMNOPpQRSTUVWXY"},
			"handles a single mixed-case duplicate character in key",
		},
	}

	for _, tc := range errorTests {
		logger.Infof("Testing %s...", tc.name)

		r := runner.Run(workDir, "substitution", tc.args...).
			WithTimeout(5 * time.Second).
			Execute().
			Exit(1)

		if err := r.Error(); err != nil {
			return fmt.Errorf("test failed for %s: %v", tc.name, err)
		}

		logger.Successf("✓ %s", tc.name)
	}

	logger.Successf("All substitution tests passed!")
	return nil
}
