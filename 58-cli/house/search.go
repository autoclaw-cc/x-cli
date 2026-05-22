package house

import (
	"58-cli/browser"
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
	u := fmt.Sprintf("https://%s.58.com/zufang/", params.City)
	if params.Page > 1 {
		u += fmt.Sprintf("pn%d/", params.Page)
	}
	if params.Keyword != "" {
		u += "?key=" + url.QueryEscape(params.Keyword)
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
		if (/验证码|安全验证|人机验证|验证/.test(text) && text.length < 500) {
			return JSON.stringify({error: "captcha_required"});
		}

		var cards = document.querySelectorAll(".house-cell");
		if (!cards.length) return JSON.stringify({error: "no_listings", message: "no .house-cell found"});

		var listings = [];
		var limit = %d;

		cards.forEach(function(c) {
			if (limit > 0 && listings.length >= limit) return;

			var desA = c.querySelector(".des a");
			var title = desA ? desA.innerText.trim().split("\n")[0] : "";
			var href = desA ? desA.href : "";

			var roomEl = c.querySelector(".room");
			var roomText = roomEl ? roomEl.innerText.trim() : "";
			var layoutMatch = roomText.match(/(\d+室)/);
			var layout = layoutMatch ? layoutMatch[1] : "";
			var areaMatch = roomText.match(/([\d.]+)\s*㎡/);
			var areaSqm = areaMatch ? parseFloat(areaMatch[1]) : 0;

			var inforEl = c.querySelector(".infor");
			var district = inforEl ? inforEl.innerText.trim().replace(/\s+/g, " ") : "";

			var moneyEl = c.querySelector(".money b");
			var priceText = moneyEl ? moneyEl.innerText.trim().replace(/,/g, "") : "0";
			var rentMonthly = parseInt(priceText) || 0;

			var jjrEl = c.querySelector(".jjr");
			var jjrText = jjrEl ? jjrEl.innerText.trim() : "";
			var posterName = "";
			var posterType = "unknown";
			var jjrMatch = jjrText.match(/来自经纪人[:：]\s*(.*)/);
			if (jjrMatch) {
				posterName = jjrMatch[1].trim();
				posterType = "agent";
			} else if (/个人/.test(jjrText)) {
				posterName = jjrText.replace(/个人/, "").trim();
				posterType = "owner";
			}

			var epLog = c.getAttribute("ep-log");
			var infoid = "";
			try { infoid = JSON.parse(epLog).infoid || ""; } catch(e) {}

			var imgEl = c.querySelector(".img-list img");
			var imgSrc = "";
			if (imgEl) {
				imgSrc = imgEl.src || imgEl.getAttribute("lazy_src") || "";
				if (imgSrc.indexOf("lazy_pic") >= 0) imgSrc = "";
			}

			var tags = [];
			c.querySelectorAll(".ax-tip, .tag").forEach(function(t) {
				var tv = t.innerText.trim();
				if (tv) tags.push(tv);
			});

			if (href) {
				listings.push({
					id: infoid,
					title: title,
					rent_monthly: rentMonthly,
					layout: layout,
					area_sqm: areaSqm,
					district: district,
					floor: "",
					tags: tags,
					poster_name: posterName,
					poster_type: posterType,
					url: href.split("?")[0],
					images: imgSrc ? [imgSrc] : []
				});
			}
		});

		var totalEl = document.querySelector(".listsum .num, .list-head .num");
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
		return nil, fmt.Errorf("58.com 安全验证触发，请在 Chrome 中完成验证后重试")
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
