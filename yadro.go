package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"yadro/config"
	"yadro/models"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error reading config %v", err)
		os.Exit(1)
	}

	_ = cfg

	file, err := os.Open("events")
	if err != nil {
		fmt.Printf("error reading events file %v", err)
		os.Exit(2)
	}

	out, err := os.Create("out.txt")
	if err != nil {
		fmt.Printf("error opening out file %v", err)
		os.Exit(3)
	}
	wr := bufio.NewWriter(out)

	competitors := make(map[string]models.Result)
	r := bufio.NewReader(file)
	for {
		var extra string
		line, err := r.ReadString('\n')
		if err == io.EOF && line == "" {
			break
		}
		line = strings.ReplaceAll(line, "\n", "")
		line = strings.ReplaceAll(line, "\r", "")
		tokens := strings.SplitN(line, " ", 4)
		timeStr := tokens[0]
		eventId := tokens[1]
		competitorId := tokens[2]
		t, _ := time.Parse("[15:04:05]", timeStr)
		if len(tokens) == 4 {
			extra = tokens[3]
		}
		switch eventId {
		case "1":
			str := fmt.Sprintf("%s The competitor(%s) registered\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
			competitors[competitorId] = models.Result{Started: false, Finished: false}
		case "2":
			str := fmt.Sprintf("%s The start time for the competitor(%s) was set by a draw %s\n", t.Format("[15:04:05.999]"), competitorId, extra)
			wr.WriteString(str)
			setTime, _ := time.Parse("15:04:05", extra)
			temp := competitors[competitorId]
			temp.SetTime = setTime
			competitors[competitorId] = temp
		case "3":
			str := fmt.Sprintf("%s The competitor(%s) is on the start line\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
		case "4":
			str := fmt.Sprintf("%s The competitor(%s) has started\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
			temp := competitors[competitorId]
			temp.Started = true
			temp.StartTime = t
			temp.StartLapTime = t
			competitors[competitorId] = temp
		case "5":
			str := fmt.Sprintf("%s The competitor(%s) is on the firing range(%s)\n", t.Format("[15:04:05.999]"), competitorId, extra)
			wr.WriteString(str)
			temp := competitors[competitorId]
			temp.ShotNumber += 5
			competitors[competitorId] = temp
		case "6":
			str := fmt.Sprintf("%s The target(%s) has been hit by competitor(%s)\n", t.Format("[15:04:05.999]"), extra, competitorId)
			wr.WriteString(str)
			temp := competitors[competitorId]
			temp.HitNumber++
			competitors[competitorId] = temp
		case "7":
			str := fmt.Sprintf("%s The competitor(%s) left firing range\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
		case "8":
			str := fmt.Sprintf("%s The competitor(%s) entered penalty laps\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
			temp := competitors[competitorId]
			temp.StartPenalty = t
			competitors[competitorId] = temp
		case "9":
			str := fmt.Sprintf("%s The competitor(%s) left penalty laps\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
			temp := competitors[competitorId]
			s := cfg.Penalty * (temp.ShotNumber - temp.HitNumber)
			zero, _ := time.Parse("15:05:04.999", "00:00:00.000")
			temp.PenaltyLapTime.LapTime = zero.Add(t.Sub(temp.StartPenalty))
			temp.PenaltyLapTime.AvgSpeed = float64(s) / float64(t.Sub(temp.StartPenalty).Seconds())
			competitors[competitorId] = temp
		case "10":
			str := fmt.Sprintf("%s The competitor(%s) ended the main lap\n", t.Format("[15:04:05.999]"), competitorId)
			wr.WriteString(str)
			zero, _ := time.Parse("15:05:04.999", "00:00:00.000")
			temp := competitors[competitorId]
			laptime := float64(t.Sub(temp.StartLapTime).Seconds())

			temp.LTAS = append(temp.LTAS, models.LapTimeAvgSpeed{LapTime: zero.Add(time.Duration(t.Sub(temp.StartLapTime))), AvgSpeed: float64(cfg.LapLen) / laptime})
			temp.StartLapTime = t
			if cfg.Laps == len(temp.LTAS) {
				temp.Finished = true
				temp.TotalTime = zero.Add(t.Sub(temp.StartTime))
			}
			competitors[competitorId] = temp
		case "11":
			str := fmt.Sprintf("%s The competitor(%s) can't continue: %s\n", t.Format("[15:04:05.999]"), competitorId, extra)
			wr.WriteString(str)
			temp := competitors[competitorId]
			i := len(temp.LTAS)
			for i < cfg.Laps {
				temp.LTAS = append(temp.LTAS, models.LapTimeAvgSpeed{})
				i++
			}
			competitors[competitorId] = temp
		}
		wr.Flush()
		if err == io.EOF {
			break
		}

	}
	sorted := make([]models.Result, len(competitors))
	for i := 0; i < len(sorted); i++ {
		strI := strconv.Itoa(i)
		sorted[i] = competitors[strI]
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].TotalTime.Format("15:04:05.000") < sorted[j].TotalTime.Format("15:04:05")
	})

	for k, v := range sorted {
		if !v.Started {
			fmt.Printf("[NotStarted] %d\n", k)
		} else if !v.Finished {
			fmt.Printf("[NotFinished] %d [", k)
			for i := 0; i < len(v.LTAS); i++ {
				if v.LTAS[i].IsEmpty() {
					fmt.Print("{,}")
				} else {
					fmt.Printf("{%s, %.3f}", v.LTAS[i].LapTime.Format("15:04:05.999"), v.LTAS[i].AvgSpeed)
				}

			}
			if v.PenaltyLapTime.IsEmpty() {
				fmt.Printf("] {,} %d/%d\n", v.HitNumber, v.ShotNumber)
			} else {
				fmt.Printf("] {%s, %.3f} %d/%d\n", v.PenaltyLapTime.LapTime.Format("15:04:05.999"), v.PenaltyLapTime.AvgSpeed, v.HitNumber, v.ShotNumber)
			}

		} else {
			fmt.Printf("[%s] %d [", v.TotalTime.Format("15:04:05.999"), k)
			for i := 0; i < len(v.LTAS); i++ {
				var formatString string
				if v.LTAS[i].IsEmpty() {
					formatString = "{,}"
				} else {
					formatString = "{%s, %.3f}"
				}
				fmt.Printf(formatString, v.LTAS[i].LapTime.Format("15:04:05.999"), v.LTAS[i].AvgSpeed)
			}
			if v.PenaltyLapTime.IsEmpty() {
				fmt.Printf("] {,} %d/%d\n", v.HitNumber, v.ShotNumber)
			} else {
				fmt.Printf("] {%s, %.3f} %d/%d\n", v.PenaltyLapTime.LapTime.Format("15:04:05.999"), v.PenaltyLapTime.AvgSpeed, v.HitNumber, v.ShotNumber)
			}
		}
	}
	defer out.Close()
	defer file.Close()

}
