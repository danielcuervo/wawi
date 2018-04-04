package timex

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"2016-04-23 12:56", true, "2016-04-23T12:56:00"},
		{"2016-04-23", true, "2016-04-23T00:00:00"},
		{"-1 day", true, now.Add(-1 * 24 * time.Hour).Format(time.RFC3339)},
		{"-20 days", true, now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10 day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-1 hour", true, now.Add(-1 * time.Hour).Format(time.RFC3339)},
		{"-10 hours", true, now.Add(-10 * time.Hour).Format(time.RFC3339)},
		{"-30 hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"-30hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"now", true, now.Format("2006-01-02T15")},
		{"", true, now.Format("2006-01-02T15")},
		{"2016/04/23", false, ""},
		{"2016/04/23 12:50", false, ""},
		{"1 day", false, ""},
		{"1 hour", false, ""},
		{"actual", false, ""},
		{strconv.FormatInt(now.Unix(), 10), true, now.Format(time.RFC3339)},
		{"wrong_timestamp", false, ""},
	}
	for _, testCase := range testCases {
		date, err := Parse(testCase.dateStr, now)
		if testCase.valid {
			assert.NoError(t, err)
			assert.Contains(t, date.Format(time.RFC3339), testCase.expected)
		} else {
			assert.Error(t, err, testCase.dateStr)
		}
	}
}

func TestParseFromDate(t *testing.T) {
	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"2016-04-23 12:56", true, "2016-04-23T12:56:00"},
		{"2016-04-23", true, "2016-04-23T00:00:00"},
		{"2016/04/23", false, ""},
		{"2016/04/23 12:50", false, ""},
	}
	for _, testCase := range testCases {
		date, err := ParseFromDate(testCase.dateStr)
		if testCase.valid {
			assert.NoError(t, err)
			assert.Contains(t, date.Format(time.RFC3339), testCase.expected)
		} else {
			assert.Error(t, err, testCase.dateStr)
		}
	}
}

func TestParseFromDaysAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"-1 day", true, now.Add(-1 * 24 * time.Hour).Format(time.RFC3339)},
		{"-20 days", true, now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10 day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"1 day", false, ""},
	}
	for _, testCase := range testCases {
		date, err := ParseFromDaysAgo(testCase.dateStr, now)
		if testCase.valid {
			assert.NoError(t, err)
			assert.Contains(t, date.Format(time.RFC3339), testCase.expected)
		} else {
			assert.Error(t, err, testCase.dateStr)
		}
	}
}

func TestParseFromHoursAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"-1 hour", true, now.Add(-1 * time.Hour).Format(time.RFC3339)},
		{"-10 hours", true, now.Add(-10 * time.Hour).Format(time.RFC3339)},
		{"-30 hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"-30hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"1 hour", false, ""},
	}
	for _, testCase := range testCases {
		date, err := ParseFromHoursAgo(testCase.dateStr, now)
		if testCase.valid {
			assert.NoError(t, err)
			assert.Contains(t, date.Format(time.RFC3339), testCase.expected)
		} else {
			assert.Error(t, err, testCase.dateStr)
		}
	}
}

func TestParseFromTimestamp(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{strconv.FormatInt(now.Unix(), 10), true, now.Format(time.RFC3339)},
		{"wrong_timestamp", false, ""},
	}
	for _, testCase := range testCases {
		date, err := ParseFromTimestamp(testCase.dateStr)
		if testCase.valid {
			assert.NoError(t, err)
			assert.Contains(t, date.Format(time.RFC3339), testCase.expected)
		} else {
			assert.Error(t, err, testCase.dateStr)
		}
	}
}
