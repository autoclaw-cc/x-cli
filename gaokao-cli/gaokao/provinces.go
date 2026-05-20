package gaokao

import (
	"fmt"
	"strings"
)

type Province struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var provinceMap = map[string]string{
	"11": "北京", "12": "天津", "13": "河北", "14": "山西", "15": "内蒙古",
	"21": "辽宁", "22": "吉林", "23": "黑龙江",
	"31": "上海", "32": "江苏", "33": "浙江", "34": "安徽", "35": "福建", "36": "江西", "37": "山东",
	"41": "河南", "42": "湖北", "43": "湖南", "44": "广东", "45": "广西", "46": "海南",
	"50": "重庆", "51": "四川", "52": "贵州", "53": "云南", "54": "西藏",
	"61": "陕西", "62": "甘肃", "63": "青海", "64": "宁夏", "65": "新疆",
}

// provinceOrder defines the display order (same as the API).
var provinceOrder = []string{
	"11", "12", "13", "14", "15",
	"21", "22", "23",
	"31", "32", "33", "34", "35", "36", "37",
	"41", "42", "43", "44", "45", "46",
	"50", "51", "52", "53", "54",
	"61", "62", "63", "64", "65",
}

// ListProvinces returns all provinces in order.
func ListProvinces() []Province {
	result := make([]Province, 0, len(provinceOrder))
	for _, id := range provinceOrder {
		result = append(result, Province{ID: id, Name: provinceMap[id]})
	}
	return result
}

// ResolveProvince takes a name or ID and returns (id, name).
// Supports exact match by ID, exact match by name, and substring match by name.
func ResolveProvince(input string) (string, string, error) {
	// Try as ID first
	if name, ok := provinceMap[input]; ok {
		return input, name, nil
	}
	// Try exact name match
	for id, name := range provinceMap {
		if name == input {
			return id, name, nil
		}
	}
	// Try substring match
	for id, name := range provinceMap {
		if strings.Contains(name, input) {
			return id, name, nil
		}
	}
	return "", "", fmt.Errorf("unknown province: %s", input)
}
