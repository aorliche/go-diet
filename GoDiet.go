package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed image/pig.png
var logoBytes []byte

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

var monthStrs = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}
	
var colorWeekdayBackground = color.Gray{Y: 60}
var colorWeekday = color.White
var colorCalendarBackground = color.Gray{Y: 30}
var colorCalendarBorder = color.Black
var colorCalendarGridStrokeNotCurrentMonth = color.Gray{Y: 70}
var colorCalendarGridStrokeCurrentMonth = color.Gray{Y: 150}
var colorDayNotCurrentMonth = color.Gray{Y: 150}
var colorDayCurrentMonth = color.White
var colorCaloriesNotCurrentMonth = color.Gray{Y: 150} //color.NRGBA{R: 255, G: 255, B: 0, A: 150}
var colorCaloriesCurrentMonth = color.White //color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var colorWeightsNotCurrentMonth = color.Gray{Y: 150} //color.NRGBA{R: 255, G: 100, B: 0, A: 150}
var colorWeightsCurrentMonth = color.White //color.NRGBA{R: 255, G: 100, B: 0, A: 255}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
}

// Custom box for days in calendar
// Can be clicked to inspect the raw entry for the day
type ClickableRectangle struct {
	widget.BaseWidget
	Rect *canvas.Rectangle
	OnTap func()
}

func NewClickableRectangle(color color.Color, onTap func()) *ClickableRectangle {
	cr := &ClickableRectangle {
		Rect: canvas.NewRectangle(color),
		OnTap: onTap,
	}
	cr.ExtendBaseWidget(cr)
	return cr
}

func (cr *ClickableRectangle) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(cr.Rect)
}

func (cr *ClickableRectangle) Tapped(ev *fyne.PointEvent) {
	if cr.OnTap != nil {
		cr.OnTap()
	}
}

// Rebuild the calendar
func createCalendarArea(
	curMonth int, 
	curYear int, 
	diaryMap map[[3]int]*Day, 
	app fyne.App) (fyne.CanvasObject, error) {

	days := make([]fyne.CanvasObject, 7)

	for i := 0; i < 7; i++ {
		rect := canvas.NewRectangle(colorWeekdayBackground)
		rect.SetMinSize(fyne.NewSize(daySideLength, dayOfWeekHeight))
		rect.StrokeColor = colorCalendarGridStrokeCurrentMonth
		rect.StrokeWidth = 2
		rect.CornerRadius = 8

		text := canvas.NewText(dayOfWeekStrs[i], colorWeekday)
		text.Alignment = fyne.TextAlignCenter

		days[i] = container.NewStack(rect, text)
	}

	daysRow := container.New(layout.NewGridLayout(7), days...)

	prevMonth := curMonth-1
	prevYear := curYear
	if prevMonth == 0 {
		prevMonth = 12
		prevYear = curYear-1
	} 

	firstDay := int(time.Date(2000 + curYear, time.Month(curMonth), 1, 0, 0, 0, 0, time.UTC).Weekday())
	prevMonthDays := daysInMonth(2000 + prevYear, time.Month(prevMonth))
	daysLimit := daysInMonth(2000 + curYear, time.Month(curMonth))
	count := 0

	calendar := make([]fyne.CanvasObject, 6*7)

	for i := 0; i < 6; i++ {
		for j := 0; j < 7; j++ {
			if i == 0 && j == firstDay {
				count = 1
			}

			month := curMonth
			day := count
			year := curYear
			var dayColor color.Color
			var caloriesColor color.Color
			var weightsColor color.Color
			var boxStrokeColor color.Color

			// Previous month
			if count == 0 {
				dayColor = colorDayNotCurrentMonth
				caloriesColor = colorCaloriesNotCurrentMonth
				weightsColor = colorWeightsNotCurrentMonth
				boxStrokeColor = colorCalendarGridStrokeNotCurrentMonth
				if month == 1 {
					month = 12
					year -= 1
				} else {
					month -= 1
				}
				day = prevMonthDays - (firstDay - j - 1)
			}
			// Next month
			if count > daysLimit {
				dayColor = colorDayNotCurrentMonth
				caloriesColor = colorCaloriesNotCurrentMonth
				weightsColor = colorWeightsNotCurrentMonth
				boxStrokeColor = colorCalendarGridStrokeNotCurrentMonth
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
				boxStrokeColor = colorCalendarGridStrokeCurrentMonth
				count++
			}

			box := NewClickableRectangle(colorCalendarBackground, func() {
				dayObj, ok := diaryMap[[3]int{month, day, year}]
				if ok {
					text := dayObj.String()
					entry :=  widget.NewMultiLineEntry()
					entry.SetText(text)

					date := "Unknown Date"
					scanner := bufio.NewScanner(strings.NewReader(text))
					if scanner.Scan() {
						date = scanner.Text()
					}
					
					entryWindow := app.NewWindow(date)
					entryWindow.SetContent(entry)
					entryWindow.Resize(fyne.NewSize(800, 300))
					entryWindow.Show()
				}
			})
			box.Rect.SetMinSize(fyne.NewSize(daySideLength, daySideLength))
			box.Rect.StrokeColor = boxStrokeColor 
			box.Rect.StrokeWidth = 2
			box.Rect.CornerRadius = 8

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

	return contentWithBorder, nil
}

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

	// For selects
	curMonth := int(time.Now().Month())
	curYear := int(time.Now().Year()) - 2000
	years := GetYears(diarySlice, true)
	yearStrs := make([]string, len(years))

	for i,year := range years {
		yearStrs[i] = strconv.Itoa(year)
	}

	calendar, _ := createCalendarArea(curMonth, curYear, diaryMap, dietApp)
	calendarContainer := container.NewStack(calendar)

	labelMonth := widget.NewLabel("Month:")
	comboMonth := widget.NewSelect(monthStrs, func(value string) {
		for i,month := range monthStrs {
			if month == value {
				curMonth = i+1
				calendar, _ := createCalendarArea(curMonth, curYear, diaryMap, dietApp)
				calendarContainer.Objects[0] = calendar
				calendarContainer.Refresh()
				break
			}
		}
	})
	comboMonth.SetSelectedIndex(curMonth-1)
	
	labelYear := widget.NewLabel("Year:")
	comboYear := widget.NewSelect(yearStrs, func(value string) {
		for _,year := range yearStrs {
			if year == value {
				curYear, _ := strconv.Atoi(year)
				curYear -= 2000
				calendar, _ := createCalendarArea(curMonth, curYear, diaryMap, dietApp)
				calendarContainer.Objects[0] = calendar
				calendarContainer.Refresh()
				break
			}
		}
	})
	comboYear.SetSelectedIndex(len(yearStrs)-1)

	back := widget.NewButtonWithIcon("Back", theme.Icon(theme.IconNameNavigateBack), func() {
		if curMonth == 1 {
			curMonth = 12
			curYear -= 1
		} else {
			curMonth -= 1
		}

		// Stop it getting confused
		yearStr := strconv.Itoa(2000+curYear)
		if !slices.Contains(yearStrs, yearStr) {
			curMonth = 1
			curYear += 1
			return
		}

		comboMonth.SetSelectedIndex(curMonth-1)
		comboYear.SetSelected(yearStr)

		calendar, _ := createCalendarArea(curMonth, curYear, diaryMap, dietApp)
		calendarContainer.Objects[0] = calendar
		calendarContainer.Refresh()
	})
	
	next := widget.NewButtonWithIcon("Next", theme.Icon(theme.IconNameNavigateNext), func() {
		if curMonth == 12 {
			curMonth = 1
			curYear += 1
		} else {
			curMonth += 1
		}

		// Stop it getting confused
		yearStr := strconv.Itoa(2000+curYear)
		if !slices.Contains(yearStrs, yearStr) {
			curMonth = 12
			curYear -= 1
			return
		}

		comboMonth.SetSelectedIndex(curMonth-1)
		comboYear.SetSelected(yearStr)

		calendar, _ := createCalendarArea(curMonth, curYear, diaryMap, dietApp)
		calendarContainer.Objects[0] = calendar
		calendarContainer.Refresh()
	})

	pigIcon := fyne.NewStaticResource("image/pig.png", logoBytes)
	calendarWindow.SetIcon(pigIcon)

	ui := container.New(
		layout.NewHBoxLayout(),
		layout.NewSpacer(),
		back, labelMonth, comboMonth, labelYear, comboYear, next,
		layout.NewSpacer(),
	)
	content := container.NewVBox(ui, calendarContainer)

	calendarWindow.SetContent(content)
	calendarWindow.ShowAndRun()
}
