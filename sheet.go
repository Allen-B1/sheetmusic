package main

import (
    "encoding/csv"
    "os"
    "errors"
)

func SheetFromId(id string) (map[string]SheetRef, error) {
    file, err := os.Open("music/" + id + "/sheet.csv")
    if err != nil {
        return nil, err
    }

    out := make(map[string]SheetRef)

	reader := csv.NewReader(file)
    record, _ := reader.Read()
    for record != nil {
        if len(record) < 2 {
            return nil, errors.New("Invalid sheet.csv file")
        }
        out[record[0]] = record[1:]
        record, _ = reader.Read()
    }

    return out, nil
}
