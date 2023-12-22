package consumer

import (
	"encoding/csv"
	"fishScraper/internal/utils"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strconv"
)

type Character struct {
	Counter prometheus.Counter
	Count   float64
}

func ReadCounts(filepath string, charCounters map[string]*Character) {
	f, err := os.Open(filepath)
	if os.IsNotExist(err) {
		println("initialization file not found")
		f, err = os.Create(filepath)
		utils.Must("create initialization file", err)

		// initialize with all characters set to 0
		w := csv.NewWriter(f)
		for name := range charCounters {
			utils.Must("write to csv file", w.Write([]string{name, "0"}))
		}
		w.Flush()
		return
	}
	utils.Must("open csv file", err)
	defer func() {
		_ = f.Close()
	}()

	// read existing file
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	utils.Must("read file", err)

	for _, record := range records {
		name := record[0]
		count, err := strconv.ParseFloat(record[1], 64)
		utils.Must("Parse float in csv", err)
		charCounters[name].Count = count
		charCounters[name].Counter.Add(count)
	}
}

func WriteCounts(filepath string, counts map[string]*Character) {
	f, err := os.Create(filepath)
	utils.Must("open counts file for writing", err)
	defer func() {
		_ = f.Close()
	}()

	w := csv.NewWriter(f)
	for name, char := range counts {
		fmt.Println("Writing ", name, ": ", char.Count)
		utils.Must("write counts", w.Write([]string{name, strconv.FormatFloat(char.Count, 'g', 10, 64)}))
	}
	w.Flush()
}
