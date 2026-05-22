package house

import (
	"anjuke-cli/browser"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Listing struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	RentMonthly int      `json:"rent_monthly"`
	Layout      string   `json:"layout"`
	AreaSqm     float64  `json:"area_sqm"`
	District    string   `json:"district"`
	Floor       string   `json:"floor"`
	Tags        []string `json:"tags,omitempty"`
	PosterName  string   `json:"poster_name"`
	PosterType  string   `json:"poster_type"`
	URL         string   `json:"url"`
	Images      []string `json:"images,omitempty"`
}

type SearchResult struct {
	Listings   []Listing `json:"listings"`
	City       string    `json:"city"`
	TotalFound string    `json:"total_found"`
	Page       int       `json:"page"`
}

type SearchParams struct {
	City     string
	Keyword  string
	MinPrice int
	MaxPrice int
	Limit    int
	Page     int
}

var CityNames = map[string]string{
	"sz": "深圳", "bj": "北京", "sh": "上海",
	"gz": "广州", "hz": "杭州", "cd": "成都",
	"tj": "天津", "nj": "南京", "wh": "武汉",
	"cs": "长沙", "cq": "重庆", "xa": "西安",
}

func Search(client *browser.Client, params SearchParams) (*SearchResult, error) {
	u := fmt.Sprintf("https://%s.zu.anjuke.com/", params.City)
	if params.Page > 1 {
		u += fmt.Sprintf("p%d/", params.Page)
	}
	if params.Keyword != "" {
		sep := "?"
		if params.Page > 1 {
			sep = "&"
		}
		u += sep + "kw=" + url.QueryEscape(params.Keyword)
	}

	if err := client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(function(){
		var text = document.body ? document.body.innerText : "";
		if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
		if (location.href.indexOf("antibot") >= 0 || location.href.indexOf("verifycode") >= 0) {
			return JSON.stringify({error: "captcha_required"});
		}
		if (/验证码|安全验证|人机验证/.test(text) && text.length < 500) {
			return JSON.stringify({error: "captcha_required"});
		}

		var cards = document.querySelectorAll(".zu-itemmod");
		if (!cards.length) return JSON.stringify({error: "no_listings", message: "no .zu-itemmod found"});

		var listings = [];
		var limit = %d;

		cards.forEach(function(c) {
			if (limit > 0 && listings.length >= limit) return;

			// Title and URL: second <a> in card (first is the image link)
			var allLinks = c.querySelectorAll("a");
			var titleA = allLinks.length >= 2 ? allLinks[1] : (allLinks[0] || null);
			var title = titleA ? titleA.innerText.trim().split("\n")[0] : "";
			var href = titleA ? titleA.href : "";

			// Extract ID from URL: /fangyuan/{id}
			var idMatch = href.match(/fangyuan\/(\d+)/);
			var id = idMatch ? idMatch[1] : "";

			// Room info: first .details-item, text like "1室0厅|35平米|高层(共17层)"
			var detailItems = c.querySelectorAll(".details-item");
			var roomText = detailItems.length >= 1 ? detailItems[0].innerText.trim() : "";
			var layoutMatch = roomText.match(/(\d+室\d*厅?)/);
			var layout = layoutMatch ? layoutMatch[1] : "";
			var areaMatch = roomText.match(/([\d.]+)\s*平/);
			var areaSqm = areaMatch ? parseFloat(areaMatch[1]) : 0;
			var floorMatch = roomText.match(/(低|中|高)层/);
			var floor = floorMatch ? floorMatch[0] : "";

			// Location: second .details-item
			var district = detailItems.length >= 2 ? detailItems[1].innerText.trim().replace(/\s+/g, " ") : "";

			// Price: .zu-side .price
			var priceEl = c.querySelector(".zu-side .price");
			var priceText = priceEl ? priceEl.innerText.trim().replace(/,/g, "") : "0";
			var rentMonthly = parseInt(priceText) || 0;

			// Tags: .bot-tag .cls-common
			var tags = [];
			c.querySelectorAll(".bot-tag .cls-common").forEach(function(t) {
				var tv = t.innerText.trim();
				if (tv) tags.push(tv);
			});

			// Agent: .detail-jjr .jjr-info
			var jjrEl = c.querySelector(".detail-jjr .jjr-info");
			var posterName = jjrEl ? jjrEl.innerText.trim() : "";
			var posterType = posterName ? "agent" : "unknown";

			// Image: first <img> in card
			var imgEl = c.querySelector("img");
			var imgSrc = "";
			if (imgEl) {
				imgSrc = imgEl.src || imgEl.getAttribute("data-src") || "";
				if (imgSrc.indexOf("data:image") === 0) imgSrc = "";
			}

			if (href) {
				listings.push({
					id: id,
					title: title,
					rent_monthly: rentMonthly,
					layout: layout,
					area_sqm: areaSqm,
					district: district,
					floor: floor,
					tags: tags,
					poster_name: posterName,
					poster_type: posterType,
					url: href.split("?")[0],
					images: imgSrc ? [imgSrc] : []
				});
			}
		});

		var totalEl = document.querySelector(".zu-glide .col-num, .list-head .num, .result-count");
		var totalFound = totalEl ? totalEl.innerText.trim() : String(listings.length);

		return JSON.stringify({
			listings: listings,
			total_found: totalFound
		});
	})()`, params.Limit)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var check struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	json.Unmarshal(raw, &check)
	if check.Error == "captcha_required" {
		return nil, fmt.Errorf("安居客安全验证触发，请在 Chrome 中完成验证后重试")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s %s", check.Error, check.Message)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result.City = params.City
	result.Page = params.Page

	return &result, nil
}
