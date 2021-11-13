package parser_test

import (
	"Deliveroo/internal/parser"
	"fmt"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

var timeRanges = []parser.TimeRange{
	parser.MinuteRange,
	parser.HourRange,
	parser.DayOfMonthRange,
	parser.MonthRange,
	parser.DayOfWeekRange,
}

func MustDataForTimeRange(tr parser.TimeRange) parser.TimeRangeData {
	if d, err := parser.DataForTimeRange(tr); err != nil {
		panic(err)
	} else {
		return d
	}
}

func TestParser_TechTestExample(t *testing.T) {
	c, err := parser.NewCronExpression("*/15 0 1,15 * 1-5 /usr/bin/find")
	require.NoError(t, err)
	require.Equal(t,
		`minute        0 15 30 45
hour          0
day of month  1 15
month         1 2 3 4 5 6 7 8 9 10 11 12
day of week   1 2 3 4 5
command       /usr/bin/find
`,
		c.String(),
	)
}

func TestCronTime_LowerBound(t *testing.T) {
	for _, tr := range timeRanges {
		ct, err := parser.NewCronTime(strconv.Itoa(MustDataForTimeRange(tr).Min), tr)
		require.NoError(t, err)
		require.Equal(t, []int{MustDataForTimeRange(tr).Min}, ct.Values())
	}
}

func TestCronTime_UpperBound(t *testing.T) {
	for _, tr := range timeRanges {
		ct, err := parser.NewCronTime(strconv.Itoa(MustDataForTimeRange(tr).Max), tr)
		require.NoError(t, err)
		require.Equal(t, []int{MustDataForTimeRange(tr).Max}, ct.Values())
	}
}

func TestCronTime_OutsideLowerBound(t *testing.T) {
	for _, tr := range timeRanges {
		_, err := parser.NewCronTime(strconv.Itoa(MustDataForTimeRange(tr).Min-1), tr)
		require.Error(t, err)
	}
}

func TestCronTime_OutsideUpperBound(t *testing.T) {
	for _, tr := range timeRanges {
		_, err := parser.NewCronTime(strconv.Itoa(MustDataForTimeRange(tr).Max+1), tr)
		require.Error(t, err)
	}
}

func TestCronTime_UnrestrictedRange(t *testing.T) {
	for _, tr := range timeRanges {
		ct, err := parser.NewCronTime("*", tr)
		var expected []int
		for i := MustDataForTimeRange(tr).Min; i <= MustDataForTimeRange(tr).Max; i++ {
			expected = append(expected, i)
		}
		require.NoError(t, err)
		require.Equal(t, expected, ct.Values())
	}
}

func TestCronTime_RangeLeftToRight(t *testing.T) {
	for _, tr := range timeRanges {
		for i := MustDataForTimeRange(tr).Min; i <= MustDataForTimeRange(tr).Max; i++ {
			rangeFmt := fmt.Sprintf("%d-%d", i, MustDataForTimeRange(tr).Max)
			var expected []int
			for j := i; j <= MustDataForTimeRange(tr).Max; j++ {
				expected = append(expected, j)
			}
			t.Run(rangeFmt, func(t *testing.T) {
				ct, err := parser.NewCronTime(rangeFmt, tr)
				require.NoError(t, err)
				require.Equal(t, expected, ct.Values())
			})
		}
	}
}

func TestCronTime_RangeRightToLeft(t *testing.T) {
	for _, tr := range timeRanges {
		for i, m := MustDataForTimeRange(tr).Max, MustDataForTimeRange(tr).Min; i >= m; i-- {
			rangeFmt := fmt.Sprintf("%d-%d", m, i)
			var expected []int
			for j := m; j <= i; j++ {
				expected = append(expected, j)
			}
			t.Run(rangeFmt, func(t *testing.T) {
				ct, err := parser.NewCronTime(rangeFmt, tr)
				require.NoError(t, err)
				require.Equal(t, expected, ct.Values())
			})
		}
	}
}

func TestCronTime_Divisor(t *testing.T) {
	for _, tr := range timeRanges {
		t.Run(string(tr), func(t *testing.T) {
			data := MustDataForTimeRange(tr)
			for i := data.Min; i <= data.Max; i++ {
				divisorFmt := fmt.Sprintf("*/%d", i)
				var expected []int
				for j := data.Min; j <= data.Max; j++ {
					if i == 0 {
						continue
					}
					if j%i == 0 {
						expected = append(expected, j)
					}
				}
				t.Run(divisorFmt, func(t *testing.T) {
					ct, err := parser.NewCronTime(divisorFmt, tr)
					require.NoError(t, err)
					require.Equal(t, expected, ct.Values())
				})
			}
		})
	}
}

func TestCronTime_MonthNames(t *testing.T) {
	months := []string{
		parser.MonthJanuary,
		parser.MonthFebruary,
		parser.MonthMarch,
		parser.MonthApril,
		parser.MonthMay,
		parser.MonthJune,
		parser.MonthJuly,
		parser.MonthAugust,
		parser.MonthSeptember,
		parser.MonthOctober,
		parser.MonthNovember,
		parser.MonthDecember,
	}
	for i, month := range months {
		ct, err := parser.NewCronTime(month, parser.MonthRange)
		require.NoError(t, err)

		vct, err := parser.NewCronTime(strconv.Itoa(i+1), parser.MonthRange)
		require.NoError(t, err)
		require.Equal(t, vct.Values(), ct.Values())
	}
}

func TestCronTime_DaysOfWeekNames(t *testing.T) {
	// sun appears twice because 0 or 7 represents Sunday
	daysOfWeek := []string{
		parser.DayOfWeekMonday,
		parser.DayOfWeekTuesday,
		parser.DayOfWeekWednesday,
		parser.DayOfWeekThursday,
		parser.DayOfWeekFriday,
		parser.DayOfWeekSaturday,
		parser.DayOfWeekSunday,
	}
	for i, day := range daysOfWeek {
		ct, err := parser.NewCronTime(day, parser.DayOfWeekRange)
		require.NoError(t, err)

		vct, err := parser.NewCronTime(strconv.Itoa(i+1), parser.DayOfWeekRange)
		require.NoError(t, err)

		if day == parser.DayOfWeekSunday {
			require.Equal(t, []int{0, 7}, ct.Values())
		} else {
			require.Equal(t, vct.Values(), ct.Values())
		}
	}
}

func TestCronTime_StepValues(t *testing.T) {
	ct, err := parser.NewCronTime("0-23/2", parser.HourRange)
	require.NoError(t, err)
	require.Equal(t, []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22}, ct.Values())
}

// Tests the examples found here: https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format
func TestCronExpression_IBMExamples(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want [][]int
	}{
		{
			"2:10 PM every Monday",
			"10 14 * * 1",
			[][]int{
				{10},
				{14},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				{1},
			},
		},
		{
			"Every day at midnight",
			"0 0 * * *",
			[][]int{
				{0},
				{0},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				{0, 1, 2, 3, 4, 5, 6, 7},
			},
		},
		{
			"Every weekday at midnight",
			"0 0 * * 1-5",
			[][]int{
				{0},
				{0},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				{1, 2, 3, 4, 5},
			},
		},
		{
			"Midnight on 1st and 15th day of the month",
			"0 0 1,15 * *",
			[][]int{
				{0},
				{0},
				{1, 15},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				{0, 1, 2, 3, 4, 5, 6, 7},
			},
		},
		{
			"6.32 PM on the 17th, 21st and 29th of November plus each Monday and Wednesday in November each year",
			"32 18 17,21,29 11 mon,wed",
			[][]int{
				{32},
				{18},
				{17, 21, 29},
				{11},
				{1, 3},
			},
		},
		{
			"multiple ranges per segment",
			"1-5,10-15 7-8,10,13-14 2-4,7,31 1-3,9-12 sun,3",
			[][]int{
				{1,2,3,4,5,10,11,12,13,14,15},
				{7,8,10,13,14},
				{2,3,4,7,31},
				{1,2,3,9,10,11,12},
				{0,3,7},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, err := parser.NewCronExpression(fmt.Sprintf("%s command", tt.arg))
			require.NoError(t, err)
			require.Equal(t, tt.want, [][]int{
				ce.Minute.Values(),
				ce.Hour.Values(),
				ce.DayOfMonth.Values(),
				ce.Month.Values(),
				ce.DayOfWeek.Values(),
			})
		})
	}
}
