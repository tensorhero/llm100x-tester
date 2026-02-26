package stages

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/hellobyte-dev/llm100x-tester/internal/helpers"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

func moviesTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "movies",
		Timeout:       60 * time.Second,
		TestFunc:      testMovies,
		RequiredFiles: []string{"1.sql", "2.sql", "3.sql", "4.sql", "5.sql", "6.sql", "7.sql", "8.sql", "9.sql", "10.sql", "11.sql", "12.sql", "13.sql"},
	}
}

func testMovies(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 打开数据库
	dbPath := filepath.Join(workDir, "movies.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open movies.db: %v", err)
	}
	defer db.Close()

	// Test 1: 2008 年电影 (无序)
	logger.Infof("Testing 1.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "1.sql", expectedMovies1); err != nil {
		return fmt.Errorf("1.sql: %v", err)
	}
	logger.Successf("✓ 1.sql produces correct result")

	// Test 2: Emma Stone 出生年份 (单值)
	logger.Infof("Testing 2.sql produces correct result...")
	if err := helpers.TestSQLSingleValue(db, workDir, "2.sql", "1988"); err != nil {
		return fmt.Errorf("2.sql: %v", err)
	}
	logger.Successf("✓ 2.sql produces correct result")

	// Test 3: 2018+ 电影按字母排序 (有序)
	logger.Infof("Testing 3.sql produces correct result...")
	if err := helpers.TestSQLSingleColOrdered(db, workDir, "3.sql", expectedMovies3); err != nil {
		return fmt.Errorf("3.sql: %v", err)
	}
	logger.Successf("✓ 3.sql produces correct result")

	// Test 4: 10.0 评分电影数量 (单值)
	logger.Infof("Testing 4.sql produces correct result...")
	if err := helpers.TestSQLSingleValue(db, workDir, "4.sql", "2"); err != nil {
		return fmt.Errorf("4.sql: %v", err)
	}
	logger.Successf("✓ 4.sql produces correct result")

	// Test 5: Harry Potter 电影 (双列有序)
	logger.Infof("Testing 5.sql produces correct result...")
	if err := helpers.TestSQLDoubleColOrdered(db, workDir, "5.sql", expectedMovies5); err != nil {
		return fmt.Errorf("5.sql: %v", err)
	}
	logger.Successf("✓ 5.sql produces correct result")

	// Test 6: 2012 年平均评分 (浮点数)
	logger.Infof("Testing 6.sql produces correct result...")
	if err := helpers.TestSQLFloat(db, workDir, "6.sql", 7.74, 0.01); err != nil {
		return fmt.Errorf("6.sql: %v", err)
	}
	logger.Successf("✓ 6.sql produces correct result")

	// Test 7: 2010 年电影及评分 (双列有序)
	logger.Infof("Testing 7.sql produces correct result...")
	if err := helpers.TestSQLDoubleColOrdered(db, workDir, "7.sql", expectedMovies7); err != nil {
		return fmt.Errorf("7.sql: %v", err)
	}
	logger.Successf("✓ 7.sql produces correct result")

	// Test 8: Toy Story 演员 (无序)
	logger.Infof("Testing 8.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "8.sql", expectedMovies8); err != nil {
		return fmt.Errorf("8.sql: %v", err)
	}
	logger.Successf("✓ 8.sql produces correct result")

	// Test 9: 2004 年电影演员按出生年份排序 (有序)
	logger.Infof("Testing 9.sql produces correct result...")
	if err := helpers.TestSQLSingleColOrdered(db, workDir, "9.sql", expectedMovies9); err != nil {
		return fmt.Errorf("9.sql: %v", err)
	}
	logger.Successf("✓ 9.sql produces correct result")

	// Test 10: 9.0+ 评分电影导演 (无序)
	logger.Infof("Testing 10.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "10.sql", expectedMovies10); err != nil {
		return fmt.Errorf("10.sql: %v", err)
	}
	logger.Successf("✓ 10.sql produces correct result")

	// Test 11: Chadwick Boseman 电影按评分排序 (有序)
	logger.Infof("Testing 11.sql produces correct result...")
	if err := helpers.TestSQLSingleColOrdered(db, workDir, "11.sql", expectedMovies11); err != nil {
		return fmt.Errorf("11.sql: %v", err)
	}
	logger.Successf("✓ 11.sql produces correct result")

	// Test 12: Johnny Depp & Helena Bonham Carter 共同电影 (无序，支持两种答案)
	logger.Infof("Testing 12.sql produces correct result...")
	if err := testMovies12(db, workDir); err != nil {
		return fmt.Errorf("12.sql: %v", err)
	}
	logger.Successf("✓ 12.sql produces correct result")

	// Test 13: Kevin Bacon 合作演员 (无序)
	logger.Infof("Testing 13.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "13.sql", expectedMovies13); err != nil {
		return fmt.Errorf("13.sql: %v", err)
	}
	logger.Successf("✓ 13.sql produces correct result")

	logger.Successf("All tests passed!")
	return nil
}

// testMovies12 handles test12's two possible answers
func testMovies12(db *sql.DB, workDir string) error {
	query, err := helpers.ReadSQLFile(workDir, "12.sql")
	if err != nil {
		return err
	}

	actual, err := helpers.ExecuteQuerySingleCol(db, query)
	if err != nil {
		return err
	}

	// Try first answer (Johnny Depp & Helena Bonham Carter)
	if helpers.EqualSets(actual, expectedMovies12a) {
		return nil
	}

	// Try second answer (Bradley Cooper & Jennifer Lawrence)
	if helpers.EqualSets(actual, expectedMovies12b) {
		return nil
	}

	return fmt.Errorf("result does not match either expected answer")
}

// 预期结果数据 (对齐 CS50 check50)

// Test 1: 2008 年电影
var expectedMovies1 = []string{
	"Iron Man",
	"The Dark Knight",
	"Slumdog Millionaire",
	"Kung Fu Panda",
}

// Test 3: 2018+ 电影按字母排序
var expectedMovies3 = []string{
	"Avengers: Infinity War",
	"Black Panther",
	"Eighth Grade",
	"Gemini Man",
	"Happy Times",
	"Incredibles 2",
	"Kirklet",
	"Ma Rainey's Black Bottom",
	"Roma",
	"The Professor",
	"Toy Story 4",
}

// Test 5: Harry Potter 电影 (title, year)
var expectedMovies5 = [][2]string{
	{"Harry Potter and the Sorcerer's Stone", "2001"},
	{"Harry Potter and the Chamber of Secrets", "2002"},
	{"Harry Potter and the Prisoner of Azkaban", "2004"},
	{"Harry Potter and the Goblet of Fire", "2005"},
	{"Harry Potter and the Order of the Phoenix", "2007"},
	{"Harry Potter and the Half-Blood Prince", "2009"},
	{"Harry Potter and the Deathly Hallows: Part 1", "2010"},
	{"Harry Potter and the Deathly Hallows: Part 2", "2011"},
	{"Harry Potter: A History of Magic", "2017"},
}

// Test 7: 2010 年电影及评分 (title, rating)
var expectedMovies7 = [][2]string{
	{"Inception", "8.8"},
	{"Toy Story 3", "8.3"},
	{"How to Train Your Dragon", "8.1"},
	{"Shutter Island", "8.1"},
	{"The King's Speech", "8.0"},
	{"Harry Potter and the Deathly Hallows: Part 1", "7.7"},
	{"Iron Man 2", "7.0"},
	{"Alice in Wonderland", "6.4"},
}

// Test 8: Toy Story 演员
var expectedMovies8 = []string{
	"Don Rickles",
	"Jim Varney",
	"Tom Hanks",
	"Tim Allen",
}

// Test 9: 2004 年电影演员按出生年份排序
var expectedMovies9 = []string{
	"Craig T. Nelson",
	"Richard Griffifths",
	"Samuel L. Jackson",
	"Holly Hunter",
	"Jason Lee",
	"Rupert Grint",
	"Daniel Radcliffe",
	"Emma Watson",
}

// Test 10: 9.0+ 评分电影导演
var expectedMovies10 = []string{
	"Christopher Nolan",
	"Frank Darabont",
	"Yimou Zhang",
}

// Test 11: Chadwick Boseman 电影按评分排序
var expectedMovies11 = []string{
	"42",
	"Black Panther",
	"Marshall",
	"Ma Rainey's Black Bottom",
	"Get on Up",
	"Draft Day",
	"Message from the King",
}

// Test 12a: Johnny Depp & Helena Bonham Carter 共同电影
var expectedMovies12a = []string{
	"Corpse Bride",
	"Charlie and the Chocolate Factory",
	"Alice in Wonderland",
	"Alice Through the Looking Glass",
}

// Test 12b: Bradley Cooper & Jennifer Lawrence 共同电影 (备选答案)
var expectedMovies12b = []string{
	"Silver Linings Playbook",
	"Serena",
	"American Hustle",
	"Joy",
}

// Test 13: Kevin Bacon 合作演员
var expectedMovies13 = []string{
	"Bill Paxton",
	"Gary Sinise",
	"James McAvoy",
	"Jennifer Lawrence",
	"Tom Cruise",
	"Michael Fassbender",
	"Tom Hanks",
}
