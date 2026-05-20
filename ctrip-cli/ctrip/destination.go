package ctrip

import (
	"ctrip-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type DestinationPOI struct {
	Name       string      `json:"name"`
	PoiID      int         `json:"poi_id"`
	CoverImage string      `json:"cover_image,omitempty"`
	Score      json.Number `json:"score,omitempty"`
}

type DestinationTab struct {
	TabType string           `json:"tab_type"`
	Title   string           `json:"title"`
	POIs    []DestinationPOI `json:"pois"`
}

type TravelNote struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type DestinationInfo struct {
	Name       string           `json:"name"`
	EName      string           `json:"e_name"`
	DistrictID int              `json:"district_id"`
	IsOverseas bool             `json:"is_overseas"`
	CoverImage string           `json:"cover_image"`
	Parent     string           `json:"parent,omitempty"`
	MustDo     []DestinationTab `json:"must_do,omitempty"`
	TravelNotes []TravelNote    `json:"travel_notes,omitempty"`
	HotelURL   string           `json:"hotel_url,omitempty"`
}

func GetDestination(client *browser.Client, destination string) (*DestinationInfo, error) {
	u := fmt.Sprintf("https://you.ctrip.com/place/%s.html", destination)

	if err := client.Navigate(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(4 * time.Second)

	js := `(function(){
		var pp = window.__NEXT_DATA__ && window.__NEXT_DATA__.props && window.__NEXT_DATA__.props.pageProps;
		if (!pp || !pp.initialState) return JSON.stringify({error: "no_data"});
		var s = pp.initialState;
		var di = s.districtInfo || {};
		var ml = s.moduleList || [];

		var info = {
			name: "",
			e_name: "",
			district_id: di.districtId || 0,
			is_overseas: di.isOverseas || false,
			cover_image: di.coverImage || "",
			parent: "",
			must_do: [],
			travel_notes: [],
			hotel_url: ""
		};

		ml.forEach(function(m){
			if (m.name === "destHead" && m.headModule) {
				var d = m.headModule.district || {};
				info.name = d.name || "";
				info.e_name = d.eName || "";
				info.cover_image = d.coverImage || info.cover_image;
				info.is_overseas = d.isOverseas || false;
				info.parent = d.parentDistrictName || "";
			}
			if (m.name === "menu" && m.menuModule) {
				var menus = m.menuModule.menuList || [];
				menus.forEach(function(menu){
					if (menu.type === 1 && menu.menuButton) {
						info.hotel_url = menu.menuButton.url || "";
					}
				});
			}
			if (m.name === "mustDo" && m.mustDoModule) {
				var tabs = m.mustDoModule.mustDoTabList || [];
				info.must_do = tabs.map(function(tab){
					return {
						tab_type: tab.tabType || "",
						title: tab.title || "",
						pois: (tab.poiList || []).slice(0, 10).map(function(p){
							return {
								name: p.name || "",
								poi_id: p.poiId || p.businessId || 0,
								cover_image: p.coverImage || "",
								score: p.commentScore || 0
							};
						})
					};
				});
			}
			if (m.name === "recommendTravel" && m.recommendTravelModule) {
				var tabList = m.recommendTravelModule.tabList || [];
				if (tabList.length > 0) {
					var notes = tabList[0].travelInfoList || [];
					info.travel_notes = notes.slice(0, 10).map(function(n){
						return {
							id: n.id || 0,
							title: n.title || "",
							author: n.author ? n.author.nickName || "" : "",
							avatar: n.author ? n.author.avatar || "" : ""
						};
					});
				}
			}
		});

		return JSON.stringify(info);
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

	var result DestinationInfo
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return &result, nil
}
