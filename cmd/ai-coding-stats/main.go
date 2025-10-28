// cmd/ai-coding-stats/main.go
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

var (
	output = flag.String("output", "csv", "输出格式: csv, json")
)

type PeriodData struct {
	Period     string `json:"period"`
	Display    string `json:"display"`
	TotalPR    int    `json:"total_pr"`
	AIPR       int    `json:"ai_pr"`
	TotalLines int    `json:"total_lines"`
	AILines    int    `json:"ai_lines"`
}

type PeriodPair struct {
	Current  PeriodData `json:"current"`
	Previous PeriodData `json:"previous"`
}

type AllPeriods struct {
	Week    PeriodPair `json:"week"`
	Month   PeriodPair `json:"month"`
	Quarter PeriodPair `json:"quarter"`
}

type RepoResult struct {
	Topic   string       `json:"topic"`
	Repo    string       `json:"repo"`
	Periods AllPeriods   `json:"periods"`
	Trend   []PeriodData `json:"trend"`
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <input.csv>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	records, err := readCSV(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input CSV: %v\n", err)
		os.Exit(1)
	}

	results, err := processRecords(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process: %v\n", err)
		os.Exit(1)
	}

	if *output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			fmt.Fprintf(os.Stderr, "JSON encode error: %v\n", err)
			os.Exit(1)
		}
	} else {
		writeCSV(os.Stdout, toCSV(results))
	}
}

func readCSV(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	return r.ReadAll()
}

func writeCSV(w *os.File, data [][]string) {
	writer := csv.NewWriter(w)
	for _, row := range data {
		writer.Write(row)
	}
	writer.Flush()
}

func processRecords(records [][]string) ([]RepoResult, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	client := github.NewClient()
	ctx := context.Background()
	now := time.Now()

	var results []RepoResult
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			continue
		}
		topic, repoURL := row[0], row[1]
		repo := parseRepoURL(repoURL)
		if repo == "" {
			continue
		}

		prs, _, err := client.GetPRList(ctx, repo, token)
		if err != nil {
			continue
		}

		periods := AllPeriods{
			Week:    calculatePeriodPair(prs, "week", now),
			Month:   calculatePeriodPair(prs, "month", now),
			Quarter: calculateQuarterPair(prs, now),
		}

		// 趋势：按周
		trendMap := make(map[string]*PeriodData)
		for _, pr := range prs {
			if pr.State != models.PRStateMerged || pr.MergedAt == "" {
				continue
			}
			t, _ := time.Parse(time.RFC3339, pr.MergedAt)
			key := getPeriodKey(t, "week")
			if key == "" {
				continue
			}
			if _, ok := trendMap[key]; !ok {
				trendMap[key] = &PeriodData{Period: key}
			}
			updateStats(trendMap[key], pr)
		}
		var trend []PeriodData
		for k, v := range trendMap {
			v.Display = formatDisplay(k, "week")
			trend = append(trend, *v)
		}

		results = append(results, RepoResult{
			Topic:   topic,
			Repo:    repo,
			Periods: periods,
			Trend:   trend,
		})
	}

	return results, nil
}

func calculatePeriodPair(prs []models.PullRequest, p string, now time.Time) PeriodPair {
	currentKey := getCurrentPeriodKey(now, p)
	previousKey := getPreviousPeriodKey(now, p)

	current := PeriodData{Period: currentKey}
	previous := PeriodData{Period: previousKey}

	for _, pr := range prs {
		if pr.State != models.PRStateMerged || pr.MergedAt == "" {
			continue
		}
		t, _ := time.Parse(time.RFC3339, pr.MergedAt)
		key := getPeriodKey(t, p)
		if key == currentKey {
			updateStats(&current, pr)
		} else if key == previousKey {
			updateStats(&previous, pr)
		}
	}

	current.Display = formatDisplay(currentKey, p)
	previous.Display = formatDisplay(previousKey, p)

	return PeriodPair{Current: current, Previous: previous}
}

func calculateQuarterPair(prs []models.PullRequest, now time.Time) PeriodPair {
	currentQ := getQuarter(now)
	previousQ, _ := getPreviousQuarter(now)

	current := PeriodData{Period: currentQ}
	previous := PeriodData{Period: previousQ}

	for _, pr := range prs {
		if pr.State != models.PRStateMerged || pr.MergedAt == "" {
			continue
		}
		t, _ := time.Parse(time.RFC3339, pr.MergedAt)
		q := getQuarter(t)
		if q == currentQ {
			updateStats(&current, pr)
		} else if q == previousQ {
			updateStats(&previous, pr)
		}
	}

	current.Display = formatQuarter(currentQ)
	previous.Display = formatQuarter(previousQ)

	return PeriodPair{Current: current, Previous: previous}
}

func updateStats(stats *PeriodData, pr models.PullRequest) {
	stats.TotalPR++
	stats.TotalLines += pr.Additions
	if pr.Author.Login == "xgopilot[bot]" {
		stats.AIPR++
		stats.AILines += pr.Additions
	}
}

func getPeriodKey(t time.Time, p string) string {
	switch p {
	case "week":
		y, w := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", y, w)
	case "month":
		return t.Format("2006-01")
	}
	return ""
}

func getCurrentPeriodKey(now time.Time, p string) string {
	return getPeriodKey(now, p)
}

func getPreviousPeriodKey(now time.Time, p string) string {
	switch p {
	case "week":
		return getPeriodKey(now.AddDate(0, 0, -7), p)
	case "month":
		return getPeriodKey(now.AddDate(0, -1, 0), p)
	}
	return ""
}

func getQuarter(t time.Time) string {
	year := t.Year()
	quarter := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%d-Q%d", year, quarter)
}

func getPreviousQuarter(now time.Time) (string, string) {
	currentQ := getQuarter(now)
	parts := strings.Split(currentQ, "-Q")
	year, _ := strconv.Atoi(parts[0])
	q, _ := strconv.Atoi(parts[1])

	prevQ := q - 1
	prevYear := year
	if prevQ == 0 {
		prevQ = 4
		prevYear--
	}
	return fmt.Sprintf("%d-Q%d", prevYear, prevQ), currentQ
}

func formatDisplay(key, p string) string {
	switch p {
	case "week":
		parts := strings.Split(key, "-W")
		if len(parts) != 2 {
			return key
		}
		year, _ := strconv.Atoi(parts[0])
		week, _ := strconv.Atoi(parts[1])
		jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
		jan4Weekday := int(jan4.Weekday())
		if jan4Weekday == 0 {
			jan4Weekday = 7
		}
		jan4Monday := jan4.AddDate(0, 0, 1-jan4Weekday)
		monday := jan4Monday.AddDate(0, 0, (week-1)*7)
		sunday := monday.AddDate(0, 0, 6)
		return fmt.Sprintf("%s ~ %s", monday.Format("1月2日"), sunday.Format("1月2日"))
	case "month":
		t, _ := time.Parse("2006-01", key)
		return t.Format("2006年1月")
	}
	return key
}

func formatQuarter(q string) string {
	parts := strings.Split(q, "-Q")
	if len(parts) != 2 {
		return q
	}
	year := parts[0]
	quarter, _ := strconv.Atoi(parts[1])
	start := (quarter-1)*3 + 1
	end := start + 2
	return fmt.Sprintf("%s年 第%d季度 (%d月 ~ %d月)", year, quarter, start, end)
}

func parseRepoURL(urlStr string) string {
	urlStr = strings.TrimSuffix(urlStr, ".git")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "github.com/")
	parts := strings.Split(urlStr, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return ""
}

func toCSV(results []RepoResult) [][]string {
	headers := []string{
		"议题名称", "仓库地址",
		"周本期", "周PR", "周AI PR", "周代码行", "周AI代码行",
		"周上期", "周PR", "周AI PR", "周代码行", "周AI代码行",
		"月本期", "月PR", "月AI PR", "月代码行", "月AI代码行",
		"月上期", "月PR", "月AI PR", "月代码行", "月AI代码行",
		"季本期", "季PR", "季AI PR", "季代码行", "季AI代码行",
		"季上期", "季PR", "季AI PR", "季代码行", "季AI代码行",
	}
	rows := [][]string{headers}

	for _, r := range results {
		rows = append(rows, []string{
			r.Topic,
			"https://github.com/" + r.Repo,
			r.Periods.Week.Current.Display, fmt.Sprintf("%d", r.Periods.Week.Current.TotalPR), fmt.Sprintf("%d", r.Periods.Week.Current.AIPR),
			fmt.Sprintf("%d", r.Periods.Week.Current.TotalLines), fmt.Sprintf("%d", r.Periods.Week.Current.AILines),
			r.Periods.Week.Previous.Display, fmt.Sprintf("%d", r.Periods.Week.Previous.TotalPR), fmt.Sprintf("%d", r.Periods.Week.Previous.AIPR),
			fmt.Sprintf("%d", r.Periods.Week.Previous.TotalLines), fmt.Sprintf("%d", r.Periods.Week.Previous.AILines),
			r.Periods.Month.Current.Display, fmt.Sprintf("%d", r.Periods.Month.Current.TotalPR), fmt.Sprintf("%d", r.Periods.Month.Current.AIPR),
			fmt.Sprintf("%d", r.Periods.Month.Current.TotalLines), fmt.Sprintf("%d", r.Periods.Month.Current.AILines),
			r.Periods.Month.Previous.Display, fmt.Sprintf("%d", r.Periods.Month.Previous.TotalPR), fmt.Sprintf("%d", r.Periods.Month.Previous.AIPR),
			fmt.Sprintf("%d", r.Periods.Month.Previous.TotalLines), fmt.Sprintf("%d", r.Periods.Month.Previous.AILines),
			r.Periods.Quarter.Current.Display, fmt.Sprintf("%d", r.Periods.Quarter.Current.TotalPR), fmt.Sprintf("%d", r.Periods.Quarter.Current.AIPR),
			fmt.Sprintf("%d", r.Periods.Quarter.Current.TotalLines), fmt.Sprintf("%d", r.Periods.Quarter.Current.AILines),
			r.Periods.Quarter.Previous.Display, fmt.Sprintf("%d", r.Periods.Quarter.Previous.TotalPR), fmt.Sprintf("%d", r.Periods.Quarter.Previous.AIPR),
			fmt.Sprintf("%d", r.Periods.Quarter.Previous.TotalLines), fmt.Sprintf("%d", r.Periods.Quarter.Previous.AILines),
		})
	}
	return rows
}
