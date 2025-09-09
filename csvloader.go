package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func LoadBoard(path string) (*Board, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.ReuseRecord = true

	byCat := map[string][]*Question{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv read: %w", err)
		}

		if len(rec) < 4 {
			return nil, fmt.Errorf("csv record has %d fields, expected at least 4", len(rec))
		}

		cat := strings.TrimSpace(rec[0])
		val, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			return nil, fmt.Errorf("bad value %q: %w", rec[1], err)
		}
		q := strings.TrimSpace(rec[2])
		a := strings.TrimSpace(rec[3])

		var imagePath string
		if len(rec) >= 5 {
			imagePath = strings.TrimSpace(rec[4])
		}

		byCat[cat] = append(byCat[cat], &Question{
			Category:  cat,
			Value:     val,
			Q:         q,
			A:         a,
			ImagePath: imagePath,
		})
	}

	if len(byCat) != ExpectedCategories {
		return nil, fmt.Errorf("expected %d categories, got %d", ExpectedCategories, len(byCat))
	}

	cats := make([]*Category, 0, ExpectedCategories)
	for cat, qs := range byCat {
		sort.Slice(qs, func(i, j int) bool { return qs[i].Value < qs[j].Value })
		if len(qs) != QuestionsPerCategory {
			return nil, fmt.Errorf("category %q has %d questions; expected %d", cat, len(qs), QuestionsPerCategory)
		}
		cats = append(cats, &Category{Name: cat, Questions: qs})
	}

	sort.Slice(cats, func(i, j int) bool { return strings.ToLower(cats[i].Name) < strings.ToLower(cats[j].Name) })

	return &Board{Categories: cats}, nil
}
