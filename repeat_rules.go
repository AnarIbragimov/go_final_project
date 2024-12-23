package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Repeat interface {
	CreateDate(now, date time.Time) (string, error)
}

type RepYear []string

func (ry RepYear) CreateDate(now, date time.Time) (string, error) {
	if date.After(now) {
		return date.Format("20060102"), nil
	}
	result := date.AddDate(1, 0, 0)
	for result.Before(now) {
		result = result.AddDate(1, 0, 0)
	}
	return result.Format("20060102"), nil
}

type RepDay []string

func (rd RepDay) CreateDate(now, date time.Time) (string, error) {
	if date.After(now) {
		return date.Format("20060102"), nil
	}
	numberOfDays, err := strconv.Atoi(rd[1])
	if err != nil {
		return "", fmt.Errorf("wrong day format: %w", err)
	}

	if numberOfDays == 1 && now.After(date) {
		return now.Format("20060102"), nil
	}

	result := date.AddDate(0, 0, numberOfDays)
	for result.Before(now) {
		result = result.AddDate(0, 0, numberOfDays)
	}
	return result.Format("20060102"), nil
}

type RepWeek []string

func (rw RepWeek) CreateDate(now, date time.Time) (string, error) {
	if date.After(now) {
		return date.Format("20060102"), nil
	}
	validDates := make([]time.Time, 0)
	var weekDays []int
	for _, weekDay := range strings.Split(rw[1], ",") {
		v, err := strconv.Atoi(weekDay)
		if err != nil || v < 1 || v > 7 {
			return "", fmt.Errorf("wrong weekday format: %s", weekDay)
		}
		weekDays = append(weekDays, v)
	}

	for _, weekDay := range weekDays {
		numberOfDays := (weekDay - int(date.Weekday()) + 7) % 7
		validDate := date.AddDate(0, 0, numberOfDays)
		for validDate.Before(now) {
			validDate = validDate.AddDate(0, 0, numberOfDays)
		}
		validDates = append(validDates, validDate)
	}

	result := validDates[0]
	for _, validDate := range validDates {
		if validDate.Before(result) {
			result = validDate
		}
	}

	return result.Format("20060102"), nil
}

type RepAnyMonth []string

func (ram RepAnyMonth) CreateDate(now, date time.Time) (string, error) {
	if date.After(now) {
		return date.Format("20060102"), nil
	}
	validDates := make([]time.Time, 0)
	var monthDays []int
	for _, monthDay := range strings.Split(ram[1], ",") {
		v, err := strconv.Atoi(monthDay)
		if err != nil || v == 0 || v < -2 || v > 31 {
			return "", fmt.Errorf("wrong monthday format: %s", monthDay)
		}
		monthDays = append(monthDays, v)
	}

	for _, monthDay := range monthDays {
		var validDate time.Time
		switch monthDay {
		case -1:
			validDate = time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		case -2:
			validDate = time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
		default:
			validDate = time.Date(date.Year(), date.Month(), monthDay, 0, 0, 0, 0, time.UTC)
		}

		for validDate.Before(now) {
			month := int(validDate.Month()) + 1
			year := validDate.Year()
			if month > 12 {
				month = 1
				year++
			}
			validDate = time.Date(year, time.Month(month), monthDay, 0, 0, 0, 0, time.UTC)
		}
		validDates = append(validDates, validDate)
	}

	result := validDates[0]
	for _, validDate := range validDates {
		if validDate.Before(result) {
			result = validDate
		}
	}

	return result.Format("20060102"), nil
}

type RepConcreteMonth []string

func (rcm RepConcreteMonth) CreateDate(now, date time.Time) (string, error) {
	if date.After(now) {
		return date.Format("20060102"), nil
	}
	validDates := make([]time.Time, 0)
	var monthDays []int
	var months []time.Month
	for _, monthDay := range strings.Split(rcm[1], ",") {
		v, err := strconv.Atoi(monthDay)
		if err != nil || v == 0 || v < -2 || v > 31 {
			return "", fmt.Errorf("wrong monthday format: %s", monthDay)
		}
		monthDays = append(monthDays, v)
	}

	for _, month := range strings.Split(rcm[2], ",") {
		v, err := strconv.Atoi(month)
		if err != nil || v < 1 || v > 12 {
			return "", fmt.Errorf("wrong month format: %s", month)
		}
		months = append(months, time.Month(v))
	}

	for _, month := range months {
		for _, monthDay := range monthDays {
			var validDate time.Time
			switch monthDay {
			case -1:
				validDate = time.Date(date.Year(), month+1, 0, 0, 0, 0, 0, time.UTC)
			case -2:
				validDate = time.Date(date.Year(), month+1, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
			default:
				validDate = time.Date(date.Year(), month, monthDay, 0, 0, 0, 0, time.UTC)
			}

			for validDate.Before(now) {
				validDate = validDate.AddDate(1, 0, 0)
			}
			validDates = append(validDates, validDate)
		}
	}

	result := validDates[0]
	for _, validDate := range validDates {
		if validDate.Before(result) {
			result = validDate
		}
	}

	return result.Format("20060102"), nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	var result string
	var dates Repeat

	if repeat == "" {
		return date, nil
	}

	referenceDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("wrong date format: %w", err)
	}

	repeatSlice := strings.Split(repeat, " ")
	switch {
	case repeatSlice[0] == "y" && len(repeatSlice) == 1:
		dates = RepYear{}
	case repeatSlice[0] == "d" && len(repeatSlice) == 2:
		dates = RepDay(repeatSlice)
	case repeatSlice[0] == "w" && len(repeatSlice) == 2:
		dates = RepWeek(repeatSlice)
	case repeatSlice[0] == "m" && len(repeatSlice) == 2:
		dates = RepAnyMonth(repeatSlice)
	case repeatSlice[0] == "m" && len(repeatSlice) == 3:
		dates = RepConcreteMonth(repeatSlice)
	default:
		return "", fmt.Errorf("wrong repeat format")
	}

	result, err = dates.CreateDate(now, referenceDate)
	if err != nil {
		return "", err
	}

	return result, nil
}
