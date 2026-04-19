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

// ExportBlock is a serializable representation of a validator in a round
type ByzantineRecord struct {
	Round      int
	Validator  int
	Byzantine  bool
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

// WriteOrderCSV writes whether each validator is Byzantine or not (each round) to a CSV file
func WriteByzantineCSV(records []ByzantineRecord, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"round", "validator_id", "byzantine"})

	for _, r := range records {
		w.Write([]string{
			itoa(r.Round),
			itoa(r.Validator),
			fmt.Sprintf("%t", r.Byzantine),
		})
	}

	return nil
}

func WriteRejectedCSV(ids []string, path string) {
    f, _ := os.Create(path)
    defer f.Close()
    w := csv.NewWriter(f)
    w.Write([]string{"block_id"})
    for _, id := range ids {
        w.Write([]string{id})
    }
    w.Flush()
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
