package model

import (
	_ "embed"

	"github.com/QuantumNous/new-api/common"
)

type GachaItemQuality struct {
	Rank   int    `json:"rank"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Emblem string `json:"emblem"`
}

type GachaItemCategory struct {
	Order int    `json:"order"`
	Name  string `json:"name"`
}

type GachaItemDesign struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Quality  string `json:"quality"`
	Icon     string `json:"icon"`
	Source   string `json:"source"`
	Use      string `json:"use"`
	Effect   string `json:"effect"`
}

type GachaItemAtlas struct {
	Version    int                          `json:"version"`
	Qualities  map[string]GachaItemQuality  `json:"qualities"`
	Categories map[string]GachaItemCategory `json:"categories"`
	Items      map[string]GachaItemDesign   `json:"items"`
}

//go:embed gacha_items.json
var gachaItemJSON []byte

var gachaItems = func() GachaItemAtlas {
	var atlas GachaItemAtlas
	if err := common.Unmarshal(gachaItemJSON, &atlas); err != nil {
		panic(err)
	}
	return atlas
}()
