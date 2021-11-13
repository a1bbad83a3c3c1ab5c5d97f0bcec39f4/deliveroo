package parser

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	Arguments = 6

	MinuteRange     TimeRange = "minute"
	HourRange       TimeRange = "hour"
	DayOfMonthRange TimeRange = "day of month"
	MonthRange      TimeRange = "month"
	DayOfWeekRange  TimeRange = "day of week"

	MonthJanuary   = "jan"
	MonthFebruary  = "feb"
	MonthMarch     = "mar"
	MonthApril     = "apr"
	MonthMay       = "may"
	MonthJune      = "jun"
	MonthJuly      = "jul"
	MonthAugust    = "aug"
	MonthSeptember = "sep"
	MonthOctober   = "oct"
	MonthNovember  = "nov"
	MonthDecember  = "dec"

	DayOfWeekMonday    = "mon"
	DayOfWeekTuesday   = "tue"
	DayOfWeekWednesday = "wed"
	DayOfWeekThursday  = "thu"
	DayOfWeekFriday    = "fri"
	DayOfWeekSaturday  = "sat"
	DayOfWeekSunday    = "sun"
)

var (
	minuteRangeData     = TimeRangeData{0, 59}
	hourRangeData       = TimeRangeData{0, 23}
	dayOfMonthRangeData = TimeRangeData{1, 31}
	monthRangeData      = TimeRangeData{1, 12}
	dayOfWeekRangeData  = TimeRangeData{0, 7}

	rangeMap = map[TimeRange]TimeRangeData{
		MinuteRange:     minuteRangeData,
		HourRange:       hourRangeData,
		DayOfMonthRange: dayOfMonthRangeData,
		MonthRange:      monthRangeData,
		DayOfWeekRange:  dayOfWeekRangeData,
	}
)

type TimeRange string

func (tr TimeRange) ParseSpecial(segment string) string {
	switch tr {
	case MonthRange:
		months := map[string]int{
			MonthJanuary:   1,
			MonthFebruary:  2,
			MonthMarch:     3,
			MonthApril:     4,
			MonthMay:       5,
			MonthJune:      6,
			MonthJuly:      7,
			MonthAugust:    8,
			MonthSeptember: 9,
			MonthOctober:   10,
			MonthNovember:  11,
			MonthDecember:  12,
		}
		if v, ok := months[segment]; ok {
			return strconv.Itoa(v)
		}
	case DayOfWeekRange:
		days := map[string]int{
			DayOfWeekMonday:    1,
			DayOfWeekTuesday:   2,
			DayOfWeekWednesday: 3,
			DayOfWeekThursday:  4,
			DayOfWeekFriday:    5,
			DayOfWeekSaturday:  6,
			DayOfWeekSunday:    7,
		}
		if v, ok := days[segment]; ok {
			// Sunday is an edge case because it can represent both 0 and 7
			if segment == DayOfWeekSunday {
				return "0,7"
			}
			return strconv.Itoa(v)
		}
	}

	return segment
}

type TimeRangeData struct {
	Min int
	Max int
}

func (tr TimeRangeData) IsWithinBounds(i int) bool {
	return i >= tr.Min && i <= tr.Max
}

func DataForTimeRange(tr TimeRange) (TimeRangeData, error) {
	if d, ok := rangeMap[tr]; ok {
		return d, nil
	}

	return TimeRangeData{}, fmt.Errorf("could not find data for time range: %v", tr)
}

type cronTime struct {
	segments  []string
	timeRange TimeRange
	values    []int
}

type valueParser func(segment string, timeRange TimeRange) (matches bool, values []int, err error)

// segmentNumberParser checks if segment is just a number
func segmentNumberParser(segment string, timeRange TimeRange) (matches bool, values []int, err error) {
	if val, err := strconv.Atoi(segment); err == nil {
		if !rangeMap[timeRange].IsWithinBounds(val) {
			return true, nil, errors.New("value for segment out of bounds")
		}
		values = append(values, val)

		return true, values, nil
	}

	return
}

// segmentRangeParser checks if segment is a range of numbers
func segmentRangeParser(segment string, timeRange TimeRange) (matches bool, values []int, err error) {
	if matches, values, _ := segmentUnrestrictedRangeParser(segment, timeRange); matches {
		return true, values, nil
	}

	split := strings.Split(segment, "-")
	if len(split) != 2 {
		return
	}

	left, err := strconv.Atoi(split[0])
	if err != nil {
		return true, nil, errors.New("left value in range for segment is invalid")
	}

	right, err := strconv.Atoi(split[1])
	if err != nil {
		return true, nil, errors.New("right value in range for segment is invalid")
	}

	if !rangeMap[timeRange].IsWithinBounds(left) || !rangeMap[timeRange].IsWithinBounds(right) {
		return true, nil, errors.New("left or right value in range for segment is out of bounds")
	}

	if right < left {
		return true, nil, errors.New("right value needs to be bigger than left value in segment")
	}

	for ; left <= right; left++ {
		values = append(values, left)
	}

	return true, values, nil
}

// segmentDivisorParser checks if segment is unrestricted range with divisor
func segmentDivisorParser(segment string, timeRange TimeRange) (matches bool, values []int, err error) {
	split := strings.Split(segment, "/")
	if len(split) != 2 {
		return
	}

	var rangeValues []int
	if matches, values, _ := segmentUnrestrictedRangeParser(split[0], timeRange); matches {
		rangeValues = values
	} else if matches, values, err := segmentRangeParser(split[0], timeRange); matches {
		if err != nil {
			return true, nil, fmt.Errorf("could not parse range: %w", err)
		}
		rangeValues = values
	} else {
		return true, nil, errors.New("left value in segment needs to be a range or unrestricted (*)")
	}

	right, err := strconv.Atoi(split[1])
	if err != nil {
		return true, nil, errors.New("right value in range for segment is invalid")
	}

	for _, i := range rangeValues {
		if right == 0 {
			continue
		}
		if i%right == 0 {
			values = append(values, i)
		}
	}

	return true, values, nil
}

// segmentUnrestrictedRangeParser checks if segment is unrestricted range identifier
func segmentUnrestrictedRangeParser(segment string, timeRange TimeRange) (matches bool, values []int, err error) {
	if segment != "*" {
		return
	}

	for i := rangeMap[timeRange].Min; i <= rangeMap[timeRange].Max; i++ {
		values = append(values, i)
	}

	return true, values, nil
}

func parseValues(segments []string, timeRange TimeRange, parsers []valueParser) ([]int, error) {
	var values []int

main:
	for i, segment := range segments {
		segment = timeRange.ParseSpecial(segment)
		for _, parser := range parsers {
			if matches, val, err := parser(segment, timeRange); matches {
				if err != nil {
					return nil, fmt.Errorf("error occurred for segment %d: %w", i, err)
				}

				values = append(values, val...)
				continue main
			}
		}

		return nil, fmt.Errorf("unknown value for segment %d", i)
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	return values, nil
}

func (ct cronTime) String() string {
	valuesAsString := make([]string, len(ct.values))
	for i, v := range ct.values {
		valuesAsString[i] = strconv.Itoa(v)
	}

	return strings.Join(valuesAsString, " ")
}

func (ct cronTime) Name() string {
	return string(ct.timeRange)
}

func (ct cronTime) Values() []int {
	return ct.values
}

type CronTime interface {
	fmt.Stringer
	Name() string
	Values() []int
}

func NewCronTime(val string, tr TimeRange) (CronTime, error) {
	var segments []string
	for _, segment := range strings.Split(val, ",") {
		segments = append(segments, strings.Split(tr.ParseSpecial(segment), ",")...)
	}

	ct := cronTime{
		segments:  segments,
		timeRange: tr,
	}

	values, err := parseValues(ct.segments, ct.timeRange, []valueParser{
		segmentUnrestrictedRangeParser,
		segmentNumberParser,

		// Divisor parser takes precedence over range parser as
		// left side of divisor can also be a range
		segmentDivisorParser,
		segmentRangeParser,
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse cron time: %w", err)
	}

	ct.values = values

	return ct, nil
}

type CronExpression struct {
	Minute     CronTime
	Hour       CronTime
	DayOfMonth CronTime
	Month      CronTime
	DayOfWeek  CronTime
	Command    string
}

func NewCronExpression(expr string) (*CronExpression, error) {
	args := strings.Split(expr, " ")
	if l := len(args); l != Arguments {
		return nil, fmt.Errorf("expected %d arguments, recieved %d", Arguments, l)
	}

	var err error
	cExpr := &CronExpression{
		Command: args[5],
	}

	if cExpr.Minute, err = NewCronTime(args[0], MinuteRange); err != nil {
		return nil, fmt.Errorf("minute value could not be parsed: %w", err)
	}

	if cExpr.Hour, err = NewCronTime(args[1], HourRange); err != nil {
		return nil, fmt.Errorf("hour value could not be parsed: %w", err)
	}

	if cExpr.DayOfMonth, err = NewCronTime(args[2], DayOfMonthRange); err != nil {
		return nil, fmt.Errorf("day of month value could not be parsed: %w", err)
	}

	if cExpr.Month, err = NewCronTime(args[3], MonthRange); err != nil {
		return nil, fmt.Errorf("month value could not be parsed: %w", err)
	}

	if cExpr.DayOfWeek, err = NewCronTime(args[4], DayOfWeekRange); err != nil {
		return nil, fmt.Errorf("day of week value could not be parsed: %w", err)
	}

	return cExpr, nil
}

func (ce *CronExpression) String() string {
	properties := []struct {
		name  string
		value interface{}
	}{
		{ce.Minute.Name(), ce.Minute},
		{ce.Hour.Name(), ce.Hour},
		{ce.DayOfMonth.Name(), ce.DayOfMonth},
		{ce.Month.Name(), ce.Month},
		{ce.DayOfWeek.Name(), ce.DayOfWeek},
		{"command", ce.Command},
	}

	output := ""
	for _, p := range properties {
		output += fmt.Sprintf("%-14s%s\n", p.name, p.value)
	}

	return output
}
