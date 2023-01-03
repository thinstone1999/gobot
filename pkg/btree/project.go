package btree

import (
	"encoding/json"
	"io/ioutil"
)

type Project struct {
	Name string      `json:"name"`
	Data TreeProject `json:"data"`
	Path string      `json:"path"`
}

type TreeProject struct {
	ID     string  `json:"id"`
	Select string  `json:"selectedTree"`
	Scope  string  `json:"scope"`
	Trees  []*Tree `json:"trees"`
}

func LoadProject(path string) (*Project, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pj Project
	err = json.Unmarshal(data, &pj)
	if err != nil {
		return nil, err
	}
	return &pj, nil
}
