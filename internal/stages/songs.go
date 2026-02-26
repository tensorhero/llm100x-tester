package stages

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/hellobyte-dev/llm100x-tester/internal/helpers"
	"github.com/hellobyte-dev/tester-utils/test_case_harness"
	"github.com/hellobyte-dev/tester-utils/tester_definition"
)

const (
	// MinReflectionWords is the minimum number of words required in answers.txt
	MinReflectionWords = 10
)

func songsTestCase() tester_definition.TestCase {
	return tester_definition.TestCase{
		Slug:          "songs",
		Timeout:       60 * time.Second,
		TestFunc:      testSongs,
		RequiredFiles: []string{"1.sql", "2.sql", "3.sql", "4.sql", "5.sql", "6.sql", "7.sql", "answers.txt"},
	}
}

func testSongs(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	workDir := harness.SubmissionDir

	// 检查 answers.txt 包含足够长的反思
	logger.Infof("Checking answers.txt reflection...")
	answersContent, err := os.ReadFile(filepath.Join(workDir, "answers.txt"))
	if err != nil {
		return fmt.Errorf("failed to read answers.txt: %v", err)
	}
	words := strings.Fields(string(answersContent))
	if len(words) < MinReflectionWords {
		return fmt.Errorf("answers.txt does not contain a sufficiently long reflection (need at least %d words, got %d)", MinReflectionWords, len(words))
	}
	logger.Successf("answers.txt reflection OK")

	// 打开数据库
	dbPath := filepath.Join(workDir, "songs.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open songs.db: %v", err)
	}
	defer db.Close()

	// 4. 运行各测试
	// Test 1: 所有歌曲名称 (无序)
	logger.Infof("Testing 1.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "1.sql", expectedSongs1); err != nil {
		return fmt.Errorf("1.sql: %v", err)
	}
	logger.Successf("✓ 1.sql produces correct result")

	// Test 2: 按 tempo 排序的歌曲名称 (有序)
	logger.Infof("Testing 2.sql produces correct result...")
	if err := helpers.TestSQLSingleColOrdered(db, workDir, "2.sql", expectedSongs2); err != nil {
		return fmt.Errorf("2.sql: %v", err)
	}
	logger.Successf("✓ 2.sql produces correct result")

	// Test 3: 前 5 首最长歌曲 (有序)
	logger.Infof("Testing 3.sql produces correct result...")
	if err := helpers.TestSQLSingleColOrdered(db, workDir, "3.sql", expectedSongs3); err != nil {
		return fmt.Errorf("3.sql: %v", err)
	}
	logger.Successf("✓ 3.sql produces correct result")

	// Test 4: 高能量歌曲 (无序)
	logger.Infof("Testing 4.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "4.sql", expectedSongs4); err != nil {
		return fmt.Errorf("4.sql: %v", err)
	}
	logger.Successf("✓ 4.sql produces correct result")

	// Test 5: 平均能量 (浮点数)
	logger.Infof("Testing 5.sql produces correct result...")
	if err := helpers.TestSQLFloat(db, workDir, "5.sql", 0.65906, 0.01); err != nil {
		return fmt.Errorf("5.sql: %v", err)
	}
	logger.Successf("✓ 5.sql produces correct result")

	// Test 6: Post Malone 的歌曲 (无序)
	logger.Infof("Testing 6.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "6.sql", expectedSongs6); err != nil {
		return fmt.Errorf("6.sql: %v", err)
	}
	logger.Successf("✓ 6.sql produces correct result")

	// Test 7: Post Malone 平均能量 (浮点数)
	logger.Infof("Testing 7.sql produces correct result...")
	if err := helpers.TestSQLFloat(db, workDir, "7.sql", 0.599, 0.01); err != nil {
		return fmt.Errorf("7.sql: %v", err)
	}
	logger.Successf("✓ 7.sql produces correct result")

	// Test 8: 含 feat. 的歌曲 (无序)
	logger.Infof("Testing 8.sql produces correct result...")
	if err := helpers.TestSQLSingleColUnordered(db, workDir, "8.sql", expectedSongs8); err != nil {
		return fmt.Errorf("8.sql: %v", err)
	}
	logger.Successf("✓ 8.sql produces correct result")

	logger.Successf("All tests passed!")
	return nil
}

// 预期结果数据 (对齐 CS50 check50)
var expectedSongs1 = []string{
	"God's Plan",
	"SAD!",
	"rockstar (feat. 21 Savage)",
	"Psycho (feat. Ty Dolla $ign)",
	"In My Feelings",
	"Better Now",
	"I Like It",
	"One Kiss (with Dua Lipa)",
	"IDGAF",
	"FRIENDS",
	"Havana",
	"Lucid Dreams",
	"Nice For What",
	"Girls Like You (feat. Cardi B)",
	"The Middle",
	"All The Stars (with SZA)",
	"no tears left to cry",
	"X",
	"Moonlight",
	"Look Alive (feat. Drake)",
	"These Days (feat. Jess Glynne, Macklemore & Dan Caplen)",
	"Te Bote - Remix",
	"Mine",
	"Youngblood",
	"New Rules",
	"Shape of You",
	"Love Lies (with Normani)",
	"Meant to Be (feat. Florida Georgia Line)",
	"Jocelyn Flores",
	"Perfect",
	"Taste (feat. Offset)",
	"Solo (feat. Demi Lovato)",
	"I Fall Apart",
	"Nevermind",
	"Echame La Culpa",
	"Eastside (with Halsey & Khalid)",
	"Never Be the Same",
	"Wolves",
	"changes",
	"In My Mind",
	"River (feat. Ed Sheeran)",
	"Dura",
	"SICKO MODE",
	"Thunder",
	"Me Niego",
	"Jackie Chan",
	"Finesse (Remix) [feat. Cardi B]",
	"Back To You - From 13 Reasons Why",
	"Let You Down",
	"Call Out My Name",
	"Ric Flair Drip (& Metro Boomin)",
	"Happier",
	"Too Good At Goodbyes",
	"Freaky Friday (feat. Chris Brown)",
	"Believer",
	"FEFE (feat. Nicki Minaj & Murda Beatz)",
	"Rise",
	"Body (feat. brando)",
	"XO TOUR Llif3",
	"Sin Pijama",
	"2002",
	"Nonstop",
	"Fuck Love (feat. Trippie Redd)",
	"In My Blood",
	"Silence",
	"God is a woman",
	"Dejala que vuelva (feat. Manuel Turizo)",
	"Flames",
	"What Lovers Do",
	"Taki Taki (with Selena Gomez, Ozuna & Cardi B)",
	"Let Me Go (with Alesso, Florida Georgia Line & watt)",
	"Feel It Still",
	"Pray For Me (with Kendrick Lamar)",
	"Walk It Talk It",
	"Him & I (with Halsey)",
	"Candy Paint",
	"Congratulations",
	"1, 2, 3 (feat. Jason Derulo & De La Ghetto)",
	"Criminal",
	"Plug Walk",
	"lovely (with Khalid)",
	"Stir Fry",
	"HUMBLE.",
	"Vaina Loca",
	"Perfect Duet (Ed Sheeran & Beyonc?)",
	"Corazon (feat. Nego do Borel)",
	"Young Dumb & Broke",
	"Siguelo Bailando",
	"Downtown",
	"Bella",
	"Promises (with Sam Smith)",
	"Yes Indeed",
	"I Like Me Better",
	"This Is Me",
	"Everybody Dies In Their Nightmares",
	"Rewrite The Stars",
	"I Miss You (feat. Julia Michaels)",
	"No Brainer",
	"Dusk Till Dawn - Radio Edit",
	"Be Alright",
}

var expectedSongs2 = []string{
	"changes",
	"SAD!",
	"God's Plan",
	"Feel It Still",
	"Criminal",
	"Lucid Dreams",
	"Him & I (with Halsey)",
	"Eastside (with Halsey & Khalid)",
	"River (feat. Ed Sheeran)",
	"In My Feelings",
	"Too Good At Goodbyes",
	"I Like Me Better",
	"These Days (feat. Jess Glynne, Macklemore & Dan Caplen)",
	"Nice For What",
	"Flames",
	"Vaina Loca",
	"Sin Pijama",
	"Bella",
	"Me Niego",
	"1, 2, 3 (feat. Jason Derulo & De La Ghetto)",
	"Plug Walk",
	"Perfect Duet (Ed Sheeran & Beyonc?)",
	"Dura",
	"Perfect",
	"FRIENDS",
	"Taki Taki (with Selena Gomez, Ozuna & Cardi B)",
	"Shape of You",
	"Echame La Culpa",
	"2002",
	"Te Bote - Remix",
	"All The Stars (with SZA)",
	"IDGAF",
	"Taste (feat. Offset)",
	"Siguelo Bailando",
	"Nevermind",
	"Ric Flair Drip (& Metro Boomin)",
	"Happier",
	"Pray For Me (with Kendrick Lamar)",
	"Back To You - From 13 Reasons Why",
	"Let Me Go (with Alesso, Florida Georgia Line & watt)",
	"Havana",
	"Solo (feat. Demi Lovato)",
	"I Miss You (feat. Julia Michaels)",
	"Finesse (Remix) [feat. Cardi B]",
	"Rise",
	"The Middle",
	"What Lovers Do",
	"lovely (with Khalid)",
	"New Rules",
	"Yes Indeed",
	"Youngblood",
	"Body (feat. brando)",
	"no tears left to cry",
	"Promises (with Sam Smith)",
	"Congratulations",
	"One Kiss (with Dua Lipa)",
	"Wolves",
	"Believer",
	"Girls Like You (feat. Cardi B)",
	"Rewrite The Stars",
	"In My Mind",
	"FEFE (feat. Nicki Minaj & Murda Beatz)",
	"Be Alright",
	"Jackie Chan",
	"Moonlight",
	"Never Be the Same",
	"Everybody Dies In Their Nightmares",
	"Fuck Love (feat. Trippie Redd)",
	"Freaky Friday (feat. Chris Brown)",
	"Jocelyn Flores",
	"Call Out My Name",
	"No Brainer",
	"I Like It",
	"Young Dumb & Broke",
	"Look Alive (feat. Drake)",
	"In My Blood",
	"Psycho (feat. Ty Dolla $ign)",
	"Silence",
	"Mine",
	"I Fall Apart",
	"Love Lies (with Normani)",
	"Better Now",
	"God is a woman",
	"Walk It Talk It",
	"Let You Down",
	"HUMBLE.",
	"Meant to Be (feat. Florida Georgia Line)",
	"Nonstop",
	"SICKO MODE",
	"XO TOUR Llif3",
	"rockstar (feat. 21 Savage)",
	"Downtown",
	"Thunder",
	"Dejala que vuelva (feat. Manuel Turizo)",
	"Candy Paint",
	"Dusk Till Dawn - Radio Edit",
	"X",
	"Stir Fry",
	"This Is Me",
	"Corazon (feat. Nego do Borel)",
}

var expectedSongs3 = []string{
	"Te Bote - Remix",
	"SICKO MODE",
	"Walk It Talk It",
	"Him & I (with Halsey)",
	"Perfect",
}

var expectedSongs4 = []string{
	"Dura",
	"Me Niego",
	"Feel It Still",
	"1, 2, 3 (feat. Jason Derulo & De La Ghetto)",
	"Criminal",
}

var expectedSongs6 = []string{
	"rockstar (feat. 21 Savage)",
	"Psycho (feat. Ty Dolla $ign)",
	"Better Now",
	"I Fall Apart",
	"Candy Paint",
	"Congratulations",
}

var expectedSongs8 = []string{
	"rockstar (feat. 21 Savage)",
	"Psycho (feat. Ty Dolla $ign)",
	"Girls Like You (feat. Cardi B)",
	"Look Alive (feat. Drake)",
	"These Days (feat. Jess Glynne, Macklemore & Dan Caplen)",
	"Meant to Be (feat. Florida Georgia Line)",
	"Taste (feat. Offset)",
	"Solo (feat. Demi Lovato)",
	"River (feat. Ed Sheeran)",
	"Finesse (Remix) [feat. Cardi B]",
	"Freaky Friday (feat. Chris Brown)",
	"FEFE (feat. Nicki Minaj & Murda Beatz)",
	"Body (feat. brando)",
	"Fuck Love (feat. Trippie Redd)",
	"Dejala que vuelva (feat. Manuel Turizo)",
	"1, 2, 3 (feat. Jason Derulo & De La Ghetto)",
	"Corazon (feat. Nego do Borel)",
	"I Miss You (feat. Julia Michaels)",
}
