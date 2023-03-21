package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/bsm/mlmetrics"
	"github.com/jonreiter/govader"
	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
)

var analyzer *govader.SentimentIntensityAnalyzer

type Info struct {
	True int
	Text string
}

func init() {
	analyzer = govader.NewSentimentIntensityAnalyzer()
}

func getInfoFromFile(fileName string) ([]Info, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var info []Info
	csvReader := csv.NewReader(file)
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		value, _ := strconv.Atoi(rec[0])

		info = append(info, Info{
			True: value,
			Text: rec[1],
		})
	}

	return info, nil
}

func getScore(txt string) govader.Sentiment {
	return analyzer.PolarityScores(txt)
}

func replace(s string) string {
	set := []string{".", ";", "...", ":", ",", "\""}
	for _, st := range set {
		s = strings.Replace(s, st, " ", 1)
	}

	return s
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func normalize(s string) string {
	// t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	// result, _, _ := transform.String(t, s)
	return godiacritics.Normalize(s)
}

func sanitizer(txt string, filters []string) string {
	if len(filters) == 0 {
		return txt
	}

	for _, ft := range filters {
		switch ft {
		case "lowercase":
			txt = toLower(txt)
		case "normalize":
			txt = normalize(txt)
		case "replace":
			txt = replace(txt)
		}
	}
	return txt
}

func printParams() {
	fmt.Println(`
	-filename:	The current .csv dataset file
	-filters:	Filters you want to appy: lowercase,replace,normalize
	`)
}

func main() {
	filenameString := flag.String("filename", "", "dataset .csv file")
	filtersString := flag.String("filters", "", "list of filters you want to apply")
	flag.Parse()

	if *filenameString == "" {
		printParams()
		os.Exit(0)
	}

	var currentFilters []string
	if *filtersString != "" {
		currentFilters = strings.Split(*filtersString, ",")
	}

	results, err := getInfoFromFile(*filenameString)
	if err != nil {
		panic(err)
	}

	var yTrue []int
	var yPred []int

	for _, r := range results {
		yTrue = append(yTrue, r.True)
	}

	// Get score and fill yPred
	for _, txt := range results {
		text := sanitizer(txt.Text, currentFilters)
		/*
			The compound score is computed by summing the valence scores of each word in the lexicon, adjusted according to the rules, and then normalized to be between -1 (most extreme negative) and +1 (most extreme positive). This is the most useful metric if you want a single unidimensional measure of sentiment for a given sentence. Calling it a 'normalized, weighted composite score' is accurate.*/
		score := getScore(strings.TrimSpace(text))

		var yPredValue int
		if score.Compound >= 0.05 {
			yPredValue = 1
		} else if score.Compound <= -0.05 {
			yPredValue = 0
		}

		yPred = append(yPred, yPredValue)
	}

	mat := mlmetrics.NewConfusionMatrix()
	for i := range yTrue {
		mat.Observe(yTrue[i], yPred[i])
	}

	// uncomment if you want to see the matrix
	// for i := 0; i < mat.Order(); i++ {
	// 	fmt.Println(mat.Row(i))
	// }

	fmt.Printf("Accuracy: %.3f\n", mat.Accuracy())
	fmt.Printf("Precision 1: %.3f\n", mat.Precision(1))
	fmt.Printf("Precision 0: %.3f\n", mat.Precision(0))
	fmt.Printf("Sensitivity 1: %.3f\n", mat.Sensitivity(1))
	fmt.Printf("Sensitivity 0: %.3f\n", mat.Sensitivity(0))
	fmt.Printf("F1 - 1: %.3f\n", mat.F1(1))
	fmt.Printf("F1 - 0: %.3f\n", mat.F1(0))
}
