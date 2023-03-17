package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jonreiter/govader"
)

var analyzer *govader.SentimentIntensityAnalyzer

type Text struct {
	Input string
}

func init() {
	analyzer = govader.NewSentimentIntensityAnalyzer()
}

func OpenAndFillTexts(fileName string) ([]Text, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var texts []Text
	csvReader := csv.NewReader(file)
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		texts = append(texts, Text{
			Input: rec[1],
		})
	}

	return texts, nil
}

func GetScore(txt string) govader.Sentiment {
	return analyzer.PolarityScores(txt)
}

func main() {
	texts, err := OpenAndFillTexts("data.csv")
	if err != nil {
		panic(err)
	}

	for _, txt := range texts {
		score := GetScore(txt.Input)
		fmt.Println("TEXT:", txt.Input)
		var res string

		res = "Setiment: "
		if score.Compound >= 0.05 {
			res += "Positive"
		} else if score.Compound > -0.05 && score.Compound < 0.05 {
			res += "Neutral"
		} else if score.Compound <= -0.05 {
			res += "Negative"
		}
		fmt.Println(res)
		fmt.Println("...................................")
	}
}
