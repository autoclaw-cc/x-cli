package booking

import (
	"booking-cli/browser"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Hotel struct {
	Name       string `json:"name"`
	Location   string `json:"location"`
	Distance   string `json:"distance"`
	Score      string `json:"score"`
	ScoreWord  string `json:"score_word"`
	Reviews    string `json:"reviews"`
	RoomType   string `json:"room_type"`
	BedType    string `json:"bed_type"`
	Price      string `json:"price"`
	OrigPrice  string `json:"orig_price,omitempty"`
	Duration   string `json:"duration"`
	FreeCancellation bool `json:"free_cancellation"`
	Deal       string `json:"deal,omitempty"`
}

type SearchResult struct {
	Hotels      []Hotel `json:"hotels"`
	Destination string  `json:"destination"`
	Checkin     string  `json:"checkin"`
	Checkout    string  `json:"checkout"`
	TotalFound  string  `json:"total_found,omitempty"`
}

func SearchHotels(client *browser.Client, destination, checkin, checkout string, adults, rooms, limit int) (*SearchResult, error) {
	if checkin == "" {
		checkin = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	}
	if checkout == "" {
		t, _ := time.Parse("2006-01-02", checkin)
		checkout = t.AddDate(0, 0, 2).Format("2006-01-02")
	}
	if adults == 0 {
		adults = 2
	}
	if rooms == 0 {
		rooms = 1
	}

	u := fmt.Sprintf("https://www.booking.com/searchresults.html?ss=%s&checkin=%s&checkout=%s&group_adults=%d&no_rooms=%d&group_children=0",
		url.QueryEscape(destination), checkin, checkout, adults, rooms)

	if err := client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(6 * time.Second)

	js := fmt.Sprintf(`(function(){
		var text = document.body ? document.body.innerText : "";
		if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
		if (text.indexOf("Verify") >= 0 && text.indexOf("email") >= 0 && text.length < 1000) {
			return JSON.stringify({error: "login_required"});
		}

		var start = text.indexOf("Sort by:");
		if (start < 0) start = text.indexOf("排序方式");
		if (start < 0) start = text.indexOf("Search results");
		if (start < 0) start = text.indexOf("搜索结果");
		if (start < 0) return JSON.stringify({error: "no_results_section"});

		var section = text.substring(start);
		var lines = section.split("\n").filter(function(l){ return l.trim(); });

		var hotels = [];
		var limit = %d;
		var i = 0;

		while (i < lines.length) {
			var l = lines[i].trim();

			// Detect hotel name: line followed by "Opens in new window"
			if (i + 1 < lines.length && (lines[i+1].trim() === "Opens in new window" || lines[i+1].trim() === "在新窗口中打开")) {
				var hotel = {
					name: l,
					location: "",
					distance: "",
					score: "",
					score_word: "",
					reviews: "",
					room_type: "",
					bed_type: "",
					price: "",
					orig_price: "",
					duration: "",
					free_cancellation: false,
					deal: ""
				};

				i += 2; // skip "Opens in new window"

				// Parse remaining fields until next hotel or end
				var fieldCount = 0;
				while (i < lines.length && fieldCount < 25) {
					var fl = lines[i].trim();
					fieldCount++;

					if (i + 1 < lines.length && (lines[i+1].trim() === "Opens in new window" || lines[i+1].trim() === "在新窗口中打开")) break;
					if (/Show on map|在地图上显示/.test(fl)) {
						hotel.location = fl.replace(/Show on map.*/, "").trim();
						var distMatch = fl.match(/(\d[\d.]+ [km]+ from centre)/);
						if (distMatch) hotel.distance = distMatch[1];
					} else if (/^Scored \d/.test(fl)) {
						// skip, the actual score is the next line
					} else if (/^\d\.\d$/.test(fl)) {
						hotel.score = fl;
					} else if (/Superb|Exceptional|Very [Gg]ood|Good|Fabulous|Wonderful|Pleasant|Review score|好极了|很棒|非常好|不错|优秀|卓越|评分/.test(fl)) {
						hotel.score_word = fl;
					} else if (/reviews?$|条评论|条住客评价/.test(fl)) {
						hotel.reviews = fl;
					} else if (/Room|Suite|Dormitory|Studio|Apartment|Bed in/.test(fl) && !hotel.room_type) {
						hotel.room_type = fl;
					} else if (/bed/.test(fl) && !hotel.bed_type) {
						hotel.bed_type = fl;
					} else if (/Free cancellation|免费取消/.test(fl)) {
						hotel.free_cancellation = true;
					} else if (/Getaway Deal|Early .* Deal|Limited-time|\u9650\u65f6\u4f18\u60e0|\u65e9\u9e1f\u4f18\u60e0/.test(fl)) {
						hotel.deal = fl;
					} else if (/^\d+ night|^\d+ \u665a/.test(fl)) {
						hotel.duration = fl;
					} else if (/^S\$|^US\$|^¥|^€|^£|^A\$|^C\$|^HK\$/.test(fl)) {
						if (!hotel.price) {
							hotel.price = fl;
						}
					} else if (/^Original price/.test(fl)) {
						var m = fl.match(/Current price (.+)\./);
						if (m) hotel.price = m[1];
						var om = fl.match(/Original price (.+?)\./);
						if (om) hotel.orig_price = om[1];
					} else if (fl === "See availability" || fl === "查看供应情况" || fl === "查看价格") {
						// End of this hotel
						break;
					}
					i++;
				}

				hotels.push(hotel);
				if (limit > 0 && hotels.length >= limit) break;
			} else {
				i++;
			}
		}

		return JSON.stringify({
			hotels: hotels,
			destination: %q,
			checkin: %q,
			checkout: %q
		});
	})()`, limit, destination, checkin, checkout)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var check struct {
		Error string `json:"error"`
	}
	json.Unmarshal(raw, &check)
	if check.Error == "login_required" {
		return nil, fmt.Errorf("Booking.com login verification detected. Please complete login in Chrome, then retry.")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s", check.Error)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return &result, nil
}
