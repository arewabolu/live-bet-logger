package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type MatchObjects struct {
	league        string
	Time          int
	homeTeam      string
	awayTeam      string
	homeTeamScore string
	awayTeamScore string
	homeWin       string
	straightDraw  string
	awayWin       string
	homeWinOrDraw string
	anyTeamWin    string
	awayWinOrDraw string
}

type MatchEvents struct {
	league        string
	Time          int
	homeTeam      string
	awayTeam      string
	homeTeamScore int
	awayTeamScore int
	homeWin       float64
	straightDraw  float64
	awayWin       float64
	homeWinOrDraw float64
	anyTeamWin    float64
	awayWinOrDraw float64
}

func (mObj MatchObjects) convertToMatchEvent() MatchEvents {
	matchEvent := MatchEvents{}

	matchEvent.league = mObj.league
	matchEvent.Time = mObj.Time
	matchEvent.homeTeam = mObj.homeTeam
	matchEvent.awayTeam = mObj.awayTeam
	var err error
	matchEvent.homeTeamScore, err = strconv.Atoi(mObj.homeTeamScore)
	if err != nil {
		panic(err)
	}
	matchEvent.awayTeamScore, err = strconv.Atoi(mObj.awayTeamScore)
	if err != nil {
		panic(err)
	}

	// convert homeWin
	if mObj.homeWin == "-" {
		matchEvent.homeWin = 0
	} else {
		val, err := strconv.ParseFloat(mObj.homeWin, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.homeWin = val
	}

	// convert straightDraw
	if mObj.straightDraw == "-" {
		matchEvent.straightDraw = 0
	} else {
		val, err := strconv.ParseFloat(mObj.straightDraw, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.straightDraw = val
	}

	// convert awayWin
	if mObj.awayWin == "-" {
		matchEvent.awayWin = 0
	} else {
		val, err := strconv.ParseFloat(mObj.awayWin, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.awayWin = val
	}

	// convert homeWinOrDraw
	if mObj.homeWinOrDraw == "-" {
		matchEvent.homeWinOrDraw = 0
	} else {
		val, err := strconv.ParseFloat(mObj.homeWinOrDraw, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.homeWinOrDraw = val
	}

	//convert anyTeamWin
	if mObj.anyTeamWin == "-" {
		matchEvent.anyTeamWin = 0
	} else {
		val, err := strconv.ParseFloat(mObj.anyTeamWin, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.anyTeamWin = val
	}

	//convert awayWinOrDraw
	if mObj.awayWinOrDraw == "-" {
		matchEvent.awayWinOrDraw = 0
	} else {
		val, err := strconv.ParseFloat(mObj.awayWinOrDraw, 64)
		if err != nil {
			panic(err)
		}
		matchEvent.awayWinOrDraw = val
	}

	return matchEvent
}

func main() {
	var sport string
	var timeout uint
	var limit int
	flag.StringVar(&sport, "s", "football", "Use to specify sport.")
	flag.UintVar(&timeout, "t", 30, "set timeout to get alerts in seconds")
	flag.IntVar(&limit, "limit", 75, "set minimum match time to filter for odds \n range is between 1 and 90")
	flag.Parse()

	ticker := time.NewTicker(5 * time.Minute)

	url := fmt.Sprintf("https://22bet.ng/en/live/%s", sport)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)

	html := visitSite(url, timeout)
	dom := createDOM(html)
	matchEvents := SeperateObjects(dom)
	logLine(matchEvents, limit, logger)

	go func() {
		for range ticker.C {
			html := visitSite(url, timeout)
			dom := createDOM(html)
			matchEvents := SeperateObjects(dom)
			logLine(matchEvents, limit, logger)
		}

	}()
	select {}
}

func clearTerm() error {
	var cmd *exec.Cmd

	// Use the appropriate clear command based on the operating system
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	}

	// Set the correct output device for the command
	cmd.Stdout = os.Stdout

	// Run the command
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func logLine(allEvents []MatchObjects, limit int, log *slog.Logger) {
	err := clearTerm()
	if err != nil {
		log.Error(err.Error())
	}
	events := validateMatchObjects(limit, allEvents)
	for _, event := range events {
		goalDiff := event.homeTeamScore - event.awayTeamScore
		switch {
		case goalDiff > 0:
			switch {
			case event.homeWin <= 1.2 && event.homeWin > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: W1:%.2f 12:%.2f\n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.homeWin, event.anyTeamWin)
				slog.Info(printOut)
			case event.homeWinOrDraw <= 1.2 && event.homeWinOrDraw > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: 1X:%.2f 12:%.2f\n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.homeWinOrDraw, event.anyTeamWin)
				slog.Info(printOut)
			}

		case goalDiff == 0:
			switch {
			case event.straightDraw <= 1.4 && event.straightDraw > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s has result:\t %d-%d\npotential bets: X:%.2f 1X:%.2f 2X:%.2f\n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.straightDraw, event.homeWinOrDraw, event.awayWinOrDraw)
				slog.Info(printOut)
			case event.homeWinOrDraw <= 1.2 && event.homeWinOrDraw > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: 1X:%.2f \n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.homeWinOrDraw)
				slog.Info(printOut)
			case event.awayWinOrDraw <= 1.2 && event.awayWinOrDraw > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: 2X:%.2f \n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.awayWinOrDraw)
				slog.Info(printOut)
			}
		case goalDiff < 0:
			switch {
			case event.awayWin <= 1.2 && event.awayWin > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: W2:%.2f 12:%.2f\n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.awayWin, event.anyTeamWin)
				slog.Info(printOut)
			case event.awayWinOrDraw <= 1.2 && event.awayWinOrDraw > 1.02:
				printOut := fmt.Sprintf("%s\n%s vs %s  has result:\t %d-%d\npotential bets: 2X:%.2f 12:%.2f\n", event.league, event.homeTeam, event.awayTeam, event.homeTeamScore, event.awayTeamScore, event.awayWinOrDraw, event.anyTeamWin)
				slog.Info(printOut)
			}
		}
	}
}

func validateMatchObjects(limit int, allEvents []MatchObjects) []MatchEvents {
	events := make([]MatchEvents, 0)
	for _, event := range allEvents {
		if event.Time < 1 {
			continue
		}
		nwEvent := event.convertToMatchEvent()
		if event.Time > limit && event.Time < 90 {
			if nwEvent.homeWin < 1.2 || nwEvent.straightDraw < 1.2 || nwEvent.awayWin < 1.2 || nwEvent.homeWinOrDraw < 1.2 || nwEvent.anyTeamWin < 1.2 || nwEvent.awayWinOrDraw < 1.2 {
				events = append(events, nwEvent)
			}
		}
	}
	return events
}

func visitSite(url string, timeout uint) string {
	var html string
	newCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	newCtx, cancel = chromedp.NewContext(newCtx)
	defer cancel()
	err := chromedp.Run(newCtx, chromedp.Navigate(url), chromedp.InnerHTML("div#sports_main_new", &html))
	if err != nil {
		log.Println(err)
	}
	return html
}

func createDOM(html string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func SeperateObjects(dom *goquery.Document) []MatchObjects {
	matches := make([]MatchObjects, 0)
	dom.Find("div.live-dashboard").Each(func(i int, s *goquery.Selection) {
		s.Find("div.dashboard").Each(func(i int, s *goquery.Selection) {
			for j := 0; j < s.Children().Size()-1; j++ {
				league := s.Find("div.c-events__item_head").Find("a.c-events__liga").Text()
				s.Find("div.c-events__item_col").Each(func(i int, s *goquery.Selection) {
					matchObj := &MatchObjects{}
					matchObj.league = league
					time := s.Find("div.c-events__time").First().Text()
					if strings.Contains(time, ":") {
						splitTime := strings.Split(time, ":")
						min, err := strconv.Atoi(splitTime[0])
						if err != nil {
							panic(err)
						}
						switch {
						case min > 0:
							matchObj.Time = min
							matchObj.homeTeam = strings.TrimSpace(s.Find("span.c-events__team").Get(0).FirstChild.Data)
							matchObj.awayTeam = strings.TrimSpace(s.Find("span.c-events__team").Get(1).FirstChild.Data)
							scores := s.Find("div.c-events__score").Find("span:not(span.c-events__fullScore)").Text()
							splitScores := strings.Split(scores, "")
							matchObj.homeTeamScore = splitScores[0]
							matchObj.awayTeamScore = splitScores[1]
							s.Find("div.c-bets").EachWithBreak(func(k int, s *goquery.Selection) bool {
								setOdd(0, &matchObj.homeWin, s)
								setOdd(1, &matchObj.straightDraw, s)
								setOdd(2, &matchObj.awayWin, s)
								setOdd(3, &matchObj.homeWinOrDraw, s)
								setOdd(4, &matchObj.anyTeamWin, s)
								setOdd(5, &matchObj.awayWinOrDraw, s)
								return k <= 6
							})
						}
					}
					matches = append(matches, *matchObj)
				})
			}
		})

	})
	return matches
}

func setOdd(index int, event *string, selector *goquery.Selection) {
	if selector.Find("div.c-bets__bet_sm").Find("span.c-bets__inner").Eq(index).Size() == 0 {
		*event = "-"
	} else {
		*event = selector.Find("div.c-bets__bet_sm").Find("span.c-bets__inner").Eq(index).Text()
	}
}
