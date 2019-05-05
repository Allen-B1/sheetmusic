package main

import (
    "encoding/json"
    "io/ioutil"
)

func SheetFromId(id string) (map[string]SheetRef, error) {
    body, err := ioutil.ReadFile("music/" + id + "/sheet.json")
    if err != nil {
        return nil, err
    }

    out := make(map[string]SheetRef)
    if err := json.Unmarshal(body, &out); err != nil {
        return nil, err
    }
    return out, nil
}
