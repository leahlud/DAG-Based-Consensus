package export

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ExportBlock is a serializable representation of a certified block
type ExportBlock struct {
	ID      string
	Round   int
	Author  int
	Parents []string
	Votes   int
}

// WriteEdgesCSV writes the DAG edges to a CSV file
func WriteEdgesCSV(blocks []ExportBlock, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"source", "target", "round", "author"})
	for _, block := range blocks {
		for _, parent := range block.Parents {
			w.Write([]string{block.ID, parent, itoa(block.Round), itoa(block.Author)})
		}
	}
	return nil
}

// WriteOrderCSV writes the total order to a CSV file
func WriteOrderCSV(order []string, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"position", "block_id"})
	for i, id := range order {
		w.Write([]string{itoa(i + 1), id})
	}
	return nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
