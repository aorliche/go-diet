// This reads my previously ad-hoc formatted diary file

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Day struct {
	DayOfWeek string
	Month int
	Day int
	Year int
	PoopStrings []string
	FoodItems []FoodItem
	NoteStrings []string
	WeightStrings []string
}

type FoodItem struct {
	Name string
	Quantity int
	Calories int
}

// Helper that should really be in the core language
func At[T any](slice []T, idx int) T {
	if idx < 0 {
		return slice[len(slice)+idx]
	}
	return slice[idx]
}

func NewDay(dayOfWeek string, month int, day int, year int) *Day {
	return &Day{
		DayOfWeek: dayOfWeek,
		Month: month,
		Day: day,
		Year: year,
		PoopStrings: make([]string, 0), 
		FoodItems: make([]FoodItem, 0),
		NoteStrings: make([]string, 0),
		WeightStrings: make([]string, 0),
	}
}

// Write day as string, formatted according to my diary file
func (day *Day) String() string {
	date := fmt.Sprintf("%s %d/%d/%d", day.DayOfWeek, day.Month, day.Day, day.Year)

	// Output food lines
	foodLines := make([]string, 0)
	foodItemStrs := make([]string, len(day.FoodItems))
	for i, item := range day.FoodItems {
		itemStr := ""
		if item.Quantity > 0 {
			itemStr += strconv.Itoa(item.Quantity) + " "
		}
		itemStr += item.Name
		if item.Calories > 0 {
			itemStr += " (" + strconv.Itoa(item.Calories) + ")"
		}
		foodItemStrs[i] = itemStr
	}
	foodLine := "- Ate: "
	for i := 0; i < len(foodItemStrs); i++ {
		if foodLine == "- Ate: " {
			foodLine += foodItemStrs[i]
		} else {
			// Cap at 110 columns
			if len(foodLine) + len(foodItemStrs[i]) + 2 >= 110 {
				foodLines = append(foodLines, foodLine)
				foodLine = "- Ate: "
				i--
				continue
			} else {
				foodLine += ", " + foodItemStrs[i]
			}
		}
	}
	if foodLine != "- Ate: " {
		foodLines = append(foodLines, foodLine)
	}

	// Add total calories
	cals := day.Calories()
	totalCalsLine := "- Total calories: " + strconv.Itoa(cals) + "-" + strconv.Itoa(cals+200)
		
	// Put everything together
	lines := make([]string, 0)
	lines = append(lines, date)
	lines = append(lines, day.PoopStrings...)
	lines = append(lines, foodLines...)
	lines = append(lines, totalCalsLine)
	lines = append(lines, day.NoteStrings...)
	lines = append(lines, day.WeightStrings...)
	
	return strings.Join(lines, "\n")
}

// Total calories for a day
func (day *Day) Calories() int {
	sum := 0
	for _, item := range day.FoodItems {
		sum += item.Calories
	}
	return sum
}

// Extract just the weight (i.e. something like 149.5) from the weight string
// We can have multiple weighing per day
func (day *Day) WeightsAsStrings() []string {
	weights := make([]string, 0)
	weightRegexp := regexp.MustCompile("^- Weight: ([0-9.]+) lbs\\.")
	for _, weightStr := range day.WeightStrings {
		match := weightRegexp.FindStringSubmatch(weightStr)
		if match != nil {
			weights = append(weights, match[1])
		}
	}
	return weights
}

// Parse a diary file formatted in my ad-hoc format
func ReadDiaryFile(filename string) ([]*Day, error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open diary file %w", err)
		return nil, err
	}
	defer file.Close()

	days := make([]*Day, 0)
	dateRegexp := regexp.MustCompile("^([A-z]+) (\\d+)/(\\d+)/(\\d+)$")
	foodRegexp1 := regexp.MustCompile("^(\\d+) ([A-z/ ]+) \\((\\d+)\\)$")
	foodRegexp2 := regexp.MustCompile("^([A-z/ ]+) \\((\\d+)\\)$")
	foodRegexp3 := regexp.MustCompile("^([A-z/ ]+)$")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 {
			continue
		}

		if line[0:1] != "-" {
			match := dateRegexp.FindStringSubmatch(line)
			if match == nil {
				err := errors.New("Bad date line format")
				return nil, fmt.Errorf("%s: %w", line, err)
			}
			m1 := match[1]
			m2, _ := strconv.Atoi(match[2])
			m3, _ := strconv.Atoi(match[3])
			m4, _ := strconv.Atoi(match[4])
			day := NewDay(m1, m2, m3, m4)
			days = append(days, day)
		} else {
			line = line[2:]
			day := At(days, -1)
			// Pooped
			if strings.HasPrefix(line, "Pooped") {
				line = "- " + line
				day.PoopStrings = append(day.PoopStrings, line)
			// Skip total calories
			} else if strings.HasPrefix(line, "Total calories: ") {
				continue
			// Food list
			} else if strings.HasPrefix(line, "Ate: ") {
				line := line[5:]
				for _, foodStr := range strings.Split(line, ", ") {
					m1 := foodRegexp1.FindStringSubmatch(foodStr)
					m2 := foodRegexp2.FindStringSubmatch(foodStr)
					m3 := foodRegexp3.FindStringSubmatch(foodStr)
					if m1 != nil {
						quant, _ := strconv.Atoi(m1[1])
						cals, _ := strconv.Atoi(m1[3])
						item := FoodItem{
							Name: m1[2],
							Quantity: quant,
							Calories: cals,
						}
						day.FoodItems = append(day.FoodItems, item)
					} else if m2 != nil {
						cals, _ := strconv.Atoi(m2[2])
						item := FoodItem{
							Name: m2[1],
							Quantity: 0,
							Calories: cals,
						}
						day.FoodItems = append(day.FoodItems, item)
					} else if m3 != nil {

					}
				}
			// Weight strings (can be more than one)
			} else if strings.HasPrefix(line, "Weight: ") {
				line = "- " + line
				day.WeightStrings = append(day.WeightStrings, line)
			// Random note (workout, work, etc.)
			} else {
				line = "- " + line
				day.NoteStrings = append(day.NoteStrings, line)
			}
		}
	}

	return days, nil
}

func DiarySliceToMap(days []*Day) map[[3]int]*Day {
	mp := make(map[[3]int]*Day)
	for _, day := range days {
		key := [3]int{day.Month, day.Day, day.Year}
		mp[key] = day
	}
	return mp 
}

/*func main() {
	diaryFilename := "/home/anton/Documents/Diary/diary.txt"
	days, err := ReadDiaryFile(diaryFilename)
	if err != nil {
		fmt.Println(err)
	}
	for _, day := range days {
		fmt.Println(day)
		fmt.Println()
	}
}*/
