package ctrip

import (
	"ctrip-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type Attraction struct {
	Name         string   `json:"name"`
	PoiID        int      `json:"poi_id"`
	Score        float64  `json:"score"`
	ReviewCount  int      `json:"review_count"`
	Price        float64  `json:"price"`
	PriceDesc    string   `json:"price_desc"`
	IsFree       bool     `json:"is_free"`
	Address      string   `json:"address,omitempty"`
	Distance     string   `json:"distance,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	ShortDesc    string   `json:"short_desc,omitempty"`
	CategoryInfo string   `json:"category_info,omitempty"`
	URL          string   `json:"url"`
	CoverImage   string   `json:"cover_image,omitempty"`
}

type AttractionSearchResult struct {
	Attractions []Attraction `json:"attractions"`
	TotalCount  int          `json:"total_count"`
	District    string       `json:"district"`
}

func SearchAttractions(client *browser.Client, destination string, limit int) (*AttractionSearchResult, error) {
	u := fmt.Sprintf("https://you.ctrip.com/sight/%s.html", destination)

	if err := client.Navigate(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(4 * time.Second)

	js := `(function(){
		var pp = window.__NEXT_DATA__ && window.__NEXT_DATA__.props && window.__NEXT_DATA__.props.pageProps;
		if (!pp || !pp.initialState) return JSON.stringify({error: "no_data"});
		var s = pp.initialState;
		var ld = s.listInitData || {};
		var al = ld.attractionList || [];
		var district = ld.districtName || "";
		var total = ld.totalCount || al.length;
		var attractions = al.map(function(a){
			var c = a.card || {};
			return {
				name: c.poiName || "",
				poi_id: c.poiId || 0,
				score: c.commentScore || 0,
				review_count: c.commentCount || 0,
				price: c.price || 0,
				price_desc: c.isFree ? "免费" : (c.priceTypeDesc ? ("¥" + c.price + " " + c.priceTypeDesc) : ""),
				is_free: c.isFree || false,
				distance: c.distanceStr || "",
				tags: (c.tagNameList || []),
				short_desc: (c.shortFeatures || [])[0] || "",
				category_info: c.sightCategoryInfo || "",
				url: c.detailUrl || "",
				cover_image: c.dynamicCoverImageUrl || c.coverImageUrl || ""
			};
		});
		return JSON.stringify({attractions: attractions, total_count: total, district: district});
	})()`

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var check struct {
		Error string `json:"error"`
	}
	json.Unmarshal(raw, &check)
	if check.Error != "" {
		return nil, fmt.Errorf("page data not available: %s", check.Error)
	}

	var result AttractionSearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if limit > 0 && len(result.Attractions) > limit {
		result.Attractions = result.Attractions[:limit]
	}

	return &result, nil
}
