package ctrip

import (
	"ctrip-cli/browser"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Hotel struct {
	Name        string    `json:"name"`
	EnName      string    `json:"en_name,omitempty"`
	Star        int       `json:"star"`
	Address     string    `json:"address"`
	Zone        string    `json:"zone"`
	City        string    `json:"city"`
	Lat         string    `json:"lat,omitempty"`
	Lng         string    `json:"lng,omitempty"`
	Score       string    `json:"score"`
	ScoreDesc   string    `json:"score_desc"`
	ReviewCount string    `json:"review_count"`
	Price       int       `json:"price"`
	Currency    string    `json:"currency"`
	PriceDesc   string    `json:"price_desc"`
	Tags        []string  `json:"tags,omitempty"`
	URL         string    `json:"url,omitempty"`
}

type HotelSearchResult struct {
	Hotels     []Hotel `json:"hotels"`
	TotalCount int     `json:"total_count"`
	PageIndex  int     `json:"page_index"`
	PageSize   int     `json:"page_size"`
}

func SearchHotels(client *browser.Client, keyword, checkin, checkout string, cityID, countryID int, limit int) (*HotelSearchResult, error) {
	if checkin == "" {
		checkin = time.Now().AddDate(0, 0, 7).Format("2006/01/02")
	}
	if checkout == "" {
		t, _ := time.Parse("2006/01/02", checkin)
		checkout = t.AddDate(0, 0, 1).Format("2006/01/02")
	}

	var u string
	if cityID > 0 {
		u = fmt.Sprintf("https://hotels.ctrip.com/hotels/list?countryId=%d&city=%d&optionId=%d&optionType=City&display=%s&checkin=%s&checkout=%s",
			countryID, cityID, cityID, url.QueryEscape(keyword), url.QueryEscape(checkin), url.QueryEscape(checkout))
	} else {
		u = fmt.Sprintf("https://hotels.ctrip.com/hotels/list?countryId=0&city=0&optionId=0&optionType=Keyword&directSearch=1&display=%s&checkin=%s&checkout=%s",
			url.QueryEscape(keyword), url.QueryEscape(checkin), url.QueryEscape(checkout))
	}

	if err := client.Navigate(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(4 * time.Second)

	js := `(function(){
		var pp = window.__NEXT_DATA__ && window.__NEXT_DATA__.props && window.__NEXT_DATA__.props.pageProps;
		if (!pp || !pp.initListData) return JSON.stringify({error: "no_data"});
		var ld = pp.initListData;
		var paging = ld.pagingInfo || {};
		var hotels = (ld.hotelList || []).map(function(h){
			var info = h.hotelInfo || {};
			var ni = info.nameInfo || {};
			var pos = info.positionInfo || {};
			var ci = info.commentInfo || {};
			var star = info.hotelStar || {};
			var room = h.roomInfo && h.roomInfo["0"] ? h.roomInfo["0"] : {};
			var pi = room.priceInfo || {};
			var coord = pos.mapCoordinate && pos.mapCoordinate[0] ? pos.mapCoordinate[0] : {};
			var rawTags = info.hotelTags; var tags = Array.isArray(rawTags) ? rawTags.map(function(t){ return t.name || t.tagName || ""; }).filter(Boolean) : [];
			return {
				name: ni.name || "",
				en_name: ni.enName || "",
				star: star.star || 0,
				address: pos.address || "",
				zone: (pos.zoneNames || [])[0] || "",
				city: pos.cityName || "",
				lat: coord.latitude || "",
				lng: coord.longitude || "",
				score: ci.commentScore || "",
				score_desc: ci.commentDescription || "",
				review_count: ci.commenterNumber || "",
				price: pi.price || 0,
				currency: pi.currency || "RMB",
				price_desc: pi.displayPrice || "",
				tags: tags
			};
		});
		return JSON.stringify({hotels: hotels, total_count: paging.totalCount || hotels.length, page_index: paging.pageIndex || 1, page_size: paging.pageSize || 10});
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

	var result HotelSearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if limit > 0 && len(result.Hotels) > limit {
		result.Hotels = result.Hotels[:limit]
	}

	return &result, nil
}
