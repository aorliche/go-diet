package main

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

const daySideLength = 120
const dayOfWeekHeight = 40

var dayOfWeekStrs = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
}

/*func createCalendarArea(month int, year int) {

}*/

func main() {
	diaryFilename := "/home/anton/Documents/Diary/diary.txt"
	diarySlice, err := ReadDiaryFile(diaryFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	diaryMap := DiarySliceToMap(diarySlice)

	dietApp := app.New()
	calendarWindow := dietApp.NewWindow("Go Diet")

	colorWeekdayBackground := color.Gray{Y: 60}
	colorWeekday := color.White
	colorCalendarBackground := color.Gray{Y: 30}
	colorCalendarBorder := color.Black
	colorDayNotCurrentMonth := color.Gray{Y: 150}
	colorDayCurrentMonth := color.White
	colorCaloriesNotCurrentMonth := color.NRGBA{R: 255, G: 255, B: 0, A: 150}
	colorCaloriesCurrentMonth := color.NRGBA{R: 255, G: 255, B: 0, A: 255}
	colorWeightsNotCurrentMonth := color.NRGBA{R: 255, G: 100, B: 0, A: 150}
	colorWeightsCurrentMonth := color.NRGBA{R: 255, G: 100, B: 0, A: 255}

	days := make([]fyne.CanvasObject, 7)

	for i := 0; i < 7; i++ {
		rect := canvas.NewRectangle(colorWeekdayBackground)
		rect.SetMinSize(fyne.NewSize(daySideLength, dayOfWeekHeight))

		text := canvas.NewText(dayOfWeekStrs[i], colorWeekday)
		text.Alignment = fyne.TextAlignCenter

		days[i] = container.NewStack(rect, text)
	}

	daysRow := container.New(layout.NewGridLayout(7), days...)

	firstDay := int(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Weekday())
	previousMonthDays := daysInMonth(2026, 8)
	daysLimit := daysInMonth(2026, 9)
	count := 0

	calendar := make([]fyne.CanvasObject, 6*7)

	for i := 0; i < 6; i++ {
		for j := 0; j < 7; j++ {
			if i == 0 && j == firstDay {
				count = 1
			}
			box := canvas.NewRectangle(colorCalendarBackground)
			box.SetMinSize(fyne.NewSize(daySideLength, daySideLength))

			month := 9
			day := count
			year := 26
			var dayColor color.Color
			var caloriesColor color.Color
			var weightsColor color.Color

			// Previous month
			if count == 0 {
				dayColor = colorDayNotCurrentMonth
				caloriesColor = colorCaloriesNotCurrentMonth
				weightsColor = colorWeightsNotCurrentMonth
				if month == 1 {
					month = 12
					year -= 1
				} else {
					month -= 1
				}
				day = previousMonthDays - (firstDay - j - 1)
			}
			// Next month
			if count > daysLimit {
				dayColor = colorDayNotCurrentMonth
				caloriesColor = colorCaloriesNotCurrentMonth
				weightsColor = colorWeightsNotCurrentMonth
				if month == 12 {
					month = 1
					year += 1
				} else {
					month += 1
				}
				day = count - daysLimit
				count++
			}
			// Current month
			if count > 0 && count <= daysLimit {
				dayColor = colorDayCurrentMonth
				caloriesColor = colorCaloriesCurrentMonth
				weightsColor = colorWeightsCurrentMonth
				count++
			}
			text := canvas.NewText(strconv.Itoa(day), dayColor)
			paddedText := container.New(
				layout.NewCustomPaddedLayout(10, 100, 10, 10),
				text,
			)
			elts := make([]fyne.CanvasObject, 2)
			elts[0] = box
			elts[1] = paddedText
			diaryDay, ok := diaryMap[[3]int{month, day, year}]
			if ok {
				calories := canvas.NewText(
					"Calories: " + strconv.Itoa(diaryDay.Calories()), 
					caloriesColor,
				)
				caloriesPadded := container.New(
					layout.NewCustomPaddedLayout(10, 60, 10, 10),
					calories,
				)
				elts = append(elts, caloriesPadded)

				weights := diaryDay.WeightsAsStrings()
				for i, weight := range weights {
					weightsText := canvas.NewText(
						weight + " lbs.",
						weightsColor,
					)
					topPad := 10
					botPad := 20
					if i == 1 {
						topPad = 30
						botPad = 0
					} else if i == 2 {
						topPad = 70
						botPad = 0
					} else if i > 2 {
						break
					}
					weightsPadded := container.New(
						layout.NewCustomPaddedLayout(float32(topPad), float32(botPad), 10, 10),
						weightsText,
					)
					elts = append(elts, weightsPadded)

				}
			} 
			calendar[i*7+j] = container.NewStack(elts...)
		}
	}
	
	grid := container.New(layout.NewGridLayout(7), calendar...)
	content := container.NewVBox(daysRow, grid)

	bg := canvas.NewRectangle(colorCalendarBorder)
	contentWithBorder := container.NewStack(bg, content)

	calendarWindow.SetContent(contentWithBorder)
	calendarWindow.ShowAndRun()
}
