package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mitchellh/go-homedir"
)

func main() {
	var folder string
	var email string

	flag.StringVar(&folder, "add", "", "add new folder to scan for git repo")
	flag.StringVar(&email, "email", "example@gmail.com", "the email to scan")

	flag.Parse()

	fmt.Println(folder)

	if folder != "" {
		scan(folder)
		return
	}

	stats(email)
}

func dumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
}

func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !slices.Contains(existing, i) {
			existing = append(existing, i)
		}
	}
	return existing
}

func openFile(filepath string) *os.File {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filepath)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Fatal(err)
			}
		}
	}

	return f
}

func parseFileLinesToSlice(filepath string) []string {
	f := openFile(filepath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}

	return lines
}

func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	dumpStringsSliceToFile(repos, filePath)
}

func gotDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gogitlocalstats"

	return dotFile
}

func scanGitFolders(folders []string, folder string) []string {
	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)

	if err != nil {
		log.Fatal(err)
	}

	files, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "./git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}

			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

func recursiveScanFolder(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

func scan(folder string) {
	fmt.Printf("Found Folder\n\n")
	folderExpanded, err := homedir.Expand(folder)
	if err != nil {
		log.Fatal(err)
	}
	//recursively scan for all directories to check for .git files
	repositories := recursiveScanFolder(folderExpanded)
	filePath := gotDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
}

// stats

const daysInLastSixMonths = 183 // todo to change 31X6
const outOfRange = 99999
const weekInLastSixMonths = 26

type column []int

// function handles formatting UI, setting text format and colors
func printCell(val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

// printing only for odd days as to keep UI clean and compact
func printDayCol(day int) {
	out := " "
	switch day {
	case 1:
		out = "Mon"
	case 3:
		out = "Wed"
	case 5:
		out = "Fri"
	}
	fmt.Print(out)
}

func printMonths() {
	// time before daysInLastSixMonths from current time
	week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
	month := week.Month()

	for {
		// checking if we switched to new month or not
		if week.Month() != month {
			fmt.Printf("%s", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf(" ")
		}
		// moving to next week
		week = week.Add(7 * time.Hour * 24)
		//checks if we have passed the current week or not,
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func printCells(cols map[int]column) {
	printMonths()
	// looping over each day [0:] -> [6:]
	for j := 6; j >= 0; j-- {
		for i := weekInLastSixMonths + 1; i >= 0; i-- { //looping over each week
			//printing all day names
			if i == weekInLastSixMonths+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				// checking if its current date and week
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					// checking if data is present for a particular day "j"
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			// if no data present for day "j" print 0
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}

// TODO: resting col for a new week not done properly
func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column) // maps week number to commits for entire week
	col := column{}              // store commits done each day for a enitre week

	for _, k := range keys {
		week := int(k / 7) // returns week number
		dayinweek := k % 7 // returns day number {0, 1, 2, 3, 4, 5, 6}

		// dayinweek = 0, represents satrting of new week reseting the col slice
		if dayinweek == 0 {
			col = column{}
		}

		col = append(col, commits[k])

		// at end of week, insert the col slice representing commits for the entire week.
		if dayinweek == 6 {
			cols[week] = col
		}
	}

	return cols
}

// storing keys in a slice and sorting it in ascending order
func sortMapIntoSlice(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlice(commits)
	cols := buildCols(keys, commits)
	printCells(cols)
}

// get begining of the day with time set to [00 hrs:00 mins:00 secs]
func getBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

// count days before from now commit was done
func countDaysSinceDate(date time.Time) int {
	days := 0
	now := getBeginningOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

// count days from present day to next sunday
func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {

	case time.Sunday:
		offset = 7

	case time.Monday:
		offset = 6

	case time.Tuesday:
		offset = 5

	case time.Wednesday:
		offset = 4

	case time.Thursday:
		offset = 3

	case time.Friday:
		offset = 2

	case time.Saturday:
		offset = 1

	}
	return offset
}

func fillCommits(email string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	// reference to header in .git
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	// getting commit history
	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	offset := calcOffset()

	// iterating over each commit in the repo log and this function computes the day ago commit was done
	// and is then adjusted with offset to align with current week

	err = iterator.ForEach(func(c *object.Commit) error {
		daysAgo := countDaysSinceDate(c.Author.When) + offset

		if c.Author.Email != email {
			return nil
		}

		if daysAgo != outOfRange {
			commits[daysAgo]++
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return commits
}

func processRepositories(email string) map[int]int {
	filePath := gotDotFilePath()
	repos := parseFileLinesToSlice(filePath)
	daysInMap := daysInLastSixMonths

	commits := make(map[int]int, daysInMap)
	for i := daysInMap; i > 0; i-- {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = fillCommits(email, path, commits)
	}

	return commits
}

func stats(email string) {
	commits := processRepositories(email)
	printCommitsStats(commits)
}
