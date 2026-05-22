package house

import (
	"anjuke-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type Poster struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Company      string `json:"company,omitempty"`
	ListingCount int    `json:"listing_count,omitempty"`
	Verified     bool   `json:"verified"`
}

type SubwayInfo struct {
	Name     string `json:"name"`
	Line     string `json:"line"`
	Distance string `json:"distance"`
}

type NearbyInfo struct {
	Subway []SubwayInfo `json:"subway,omitempty"`
}

type ListingDetail struct {
	ID           string     `json:"id"`
	Platform     string     `json:"platform"`
	Title        string     `json:"title"`
	RentMonthly  int        `json:"rent_monthly"`
	DepositModel string     `json:"deposit_model,omitempty"`
	RentType     string     `json:"rent_type,omitempty"`
	Layout       string     `json:"layout"`
	AreaSqm      float64    `json:"area_sqm"`
	Floor        string     `json:"floor,omitempty"`
	Orientation  string     `json:"orientation,omitempty"`
	Decoration   string     `json:"decoration,omitempty"`
	District     string     `json:"district,omitempty"`
	Address      string     `json:"address,omitempty"`
	Compound     string     `json:"compound,omitempty"`
	BizCircle    string     `json:"biz_circle,omitempty"`
	BuiltYear    string     `json:"built_year,omitempty"`
	Poster       Poster     `json:"poster"`
	Nearby       NearbyInfo `json:"nearby"`
	Features     []string   `json:"features,omitempty"`
	Description  string     `json:"description,omitempty"`
	Images       []string   `json:"images,omitempty"`
	URL          string     `json:"url"`
}

func GetDetail(client *browser.Client, detailURL string) (*ListingDetail, error) {
	if err := client.NavigateNewTab(detailURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	time.Sleep(5 * time.Second)

	js := `(function(){
		var text = document.body ? document.body.innerText : "";
		if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
		if (location.href.indexOf("antibot") >= 0 || location.href.indexOf("verifycode") >= 0) {
			return JSON.stringify({error: "captcha_required"});
		}
		if (/验证码|安全验证|人机验证/.test(text) && text.length < 500) {
			return JSON.stringify({error: "captcha_required"});
		}

		var r = {};
		var m;

		// title
		var titleEl = document.querySelector(".house-title h3, .house-title");
		r.title = titleEl ? titleEl.innerText.trim().split("\n")[0] : "";
		if (!r.title) {
			var lines = text.split("\n").filter(function(l){ return l.trim().length > 0; });
			for (var i = 0; i < Math.min(lines.length, 15); i++) {
				if (lines[i].indexOf("租房") >= 0 && lines[i].indexOf(">") >= 0) {
					r.title = (i + 1 < lines.length) ? lines[i + 1].trim() : "";
					break;
				}
			}
		}

		// price
		m = text.match(/([\d,]+)\s*元\/月/) || text.match(/([\d,]+)\s*元\s*\/\s*月/);
		r.rent_monthly = m ? parseInt(m[1].replace(/,/g, "")) : 0;
		if (!r.rent_monthly) {
			var priceEl = document.querySelector(".price .num, .price b, .rent-price");
			if (priceEl) r.rent_monthly = parseInt(priceEl.innerText.trim().replace(/,/g, "")) || 0;
		}

		// deposit model
		m = text.match(/押[一二三]付[一二三]|半年付|年付|押一付一|面议/);
		r.deposit_model = m ? m[0] : "";

		// rent type
		m = text.match(/租赁方式[：:]\s*(\S+)/) || text.match(/(整租|合租)/);
		r.rent_type = m ? m[1] : "";

		// layout, area, decoration
		m = text.match(/户\s*型[：:]\s*([^\n]+)/) || text.match(/房屋类型[：:]\s*([^\n]+)/);
		var layoutLine = m ? m[1].trim() : "";
		r.layout = "";
		r.area_sqm = 0;
		r.decoration = "";
		if (layoutLine) {
			var lm = layoutLine.match(/(\d+室\d*厅?\d*卫?)/);
			r.layout = lm ? lm[1] : layoutLine.split(/\s/)[0];
		}
		m = text.match(/面\s*积[：:]\s*([\d.]+)/) || text.match(/([\d.]+)\s*[平㎡]/);
		r.area_sqm = m ? parseFloat(m[1]) : 0;
		if (/精装/.test(text)) r.decoration = "精装";
		else if (/简装/.test(text)) r.decoration = "简装";
		else if (/毛坯/.test(text)) r.decoration = "毛坯";
		else if (/豪装/.test(text)) r.decoration = "豪装";

		// floor and orientation
		m = text.match(/楼\s*层[：:]\s*([^\n]+)/) || text.match(/朝向楼层[：:]\s*([^\n]+)/);
		r.floor = "";
		r.orientation = "";
		if (m) {
			var floorLine = m[1].trim();
			var fm = floorLine.match(/(低|中|高)层[/／(（]\s*共?(\d+)层/);
			r.floor = fm ? fm[1] + "层/共" + fm[2] + "层" : floorLine.split(/\s/)[0];
		}
		m = text.match(/朝\s*向[：:]\s*(\S+)/) || text.match(/(朝[东南西北]+)/);
		r.orientation = m ? m[1] : "";

		// compound
		m = text.match(/小\s*区[：:]\s*([^(（\n]+)/) || text.match(/所在小区[：:]\s*([^(（\n]+)/);
		r.compound = m ? m[1].trim() : "";

		// district
		m = text.match(/区\s*域[：:]\s*([^\n]+)/) || text.match(/所属区域[：:]\s*([^\n]+)/);
		r.district = m ? m[1].trim().replace(/\s+/g, " ") : "";

		// address
		m = text.match(/地\s*址[：:]\s*([^\n]+)/) || text.match(/详细地址[：:]\s*([^\n]+)/);
		r.address = m ? m[1].trim().replace(/\s*查看地图.*/, "").replace(/\s*附近.*/, "") : "";

		// biz circle
		m = text.match(/商\s*圈[：:]\s*([^\n]+)/);
		r.biz_circle = m ? m[1].trim() : "";

		// built year
		m = text.match(/建筑年代[：:]\s*(\d+)/) || text.match(/(\d{4})年建/);
		r.built_year = m ? m[1] : "";

		// subway info
		r.nearby = {subway: []};
		var subwayRe = /地铁(\d+号线)\s*[-—·]?\s*([^\d\s]+?)站?\s*(\d+)\s*米/g;
		var sm;
		while ((sm = subwayRe.exec(text)) !== null) {
			r.nearby.subway.push({
				line: sm[1],
				name: sm[2].replace(/站$/, ""),
				distance: sm[3] + "m"
			});
			if (r.nearby.subway.length >= 5) break;
		}

		// agent
		r.poster = {name: "", type: "unknown", company: "", listing_count: 0, verified: false};
		m = text.match(/(\S+)\s*[（(]经纪人[）)]/);
		if (m) {
			r.poster.name = m[1];
			r.poster.type = "agent";
			var cm = text.match(/[（(]经纪人[）)]\s*\n\s*([^\n]+)/);
			r.poster.company = cm ? cm[1].trim() : "";
		} else {
			m = text.match(/经纪人[：:]\s*(\S+)/);
			if (m) {
				r.poster.name = m[1];
				r.poster.type = "agent";
			} else {
				m = text.match(/(\S+)\s*[（(]品牌公寓[）)]/);
				if (m) {
					r.poster.name = m[1];
					r.poster.type = "brand";
				} else {
					m = text.match(/(\S+)\s*[（(]个人[）)]/);
					if (m) {
						r.poster.name = m[1];
						r.poster.type = "owner";
					}
				}
			}
		}
		// Try agent name from DOM
		if (!r.poster.name) {
			var agentEl = document.querySelector(".broker-name, .jjr-name, .agent-name");
			if (agentEl) {
				r.poster.name = agentEl.innerText.trim();
				r.poster.type = "agent";
			}
		}
		// Company from DOM
		if (!r.poster.company) {
			var companyEl = document.querySelector(".broker-company, .jjr-company, .agent-company");
			if (companyEl) r.poster.company = companyEl.innerText.trim();
		}

		// verified
		if (/实名认证|已认证/.test(text)) r.poster.verified = true;
		if (/营业执照/.test(text)) r.poster.verified = true;

		// listing count
		m = text.match(/在租\s*(\d+)\s*套/) || text.match(/(\d+)\s*\n\s*在租房源/);
		r.poster.listing_count = m ? parseInt(m[1]) : 0;

		// id from URL
		var idMatch = location.href.match(/fangyuan\/(\d+)/);
		r.id = idMatch ? idMatch[1] : "";
		// Also try from page
		if (!r.id) {
			m = text.match(/房源编[号码][：:]\s*(\d+)/) || text.match(/编号[：:]\s*(\d+)/);
			r.id = m ? m[1] : "";
		}

		// features
		r.features = [];
		var featureEls = document.querySelectorAll(".tag-list .item, .highlight .item, .house-label span");
		if (featureEls.length > 0) {
			featureEls.forEach(function(el) {
				var ft = el.innerText.trim();
				if (ft) r.features.push(ft);
			});
		}
		if (r.features.length === 0) {
			m = text.match(/房屋亮点\s*\n([^\n]+)/) || text.match(/房源亮点\s*\n([^\n]+)/);
			r.features = m ? m[1].trim().split(/\s+/) : [];
		}

		// description
		m = text.match(/房源描述\s*\n([\s\S]*?)(?=\n房源编|周边|猜你喜欢|免责|举报|$)/);
		if (!m) m = text.match(/房屋描述\s*\n([\s\S]*?)(?=\n房源编|周边|猜你喜欢|免责|举报|$)/);
		r.description = m ? m[1].trim().substring(0, 1000) : "";

		// images
		var imgs = [];
		document.querySelectorAll("img").forEach(function(el) {
			var s = el.src || el.getAttribute("data-src") || "";
			if (s && s.indexOf("http") === 0 && s.indexOf("logo") < 0 &&
				s.indexOf("avatar") < 0 && s.indexOf("icon") < 0 &&
				s.indexOf("defaultImg") < 0 && s.indexOf("load-img") < 0 &&
				s.indexOf("deft-img") < 0 &&
				(s.indexOf("ajkimg") >= 0 || s.indexOf("anjukestatic") >= 0 || s.indexOf("aifang") >= 0) &&
				imgs.indexOf(s) < 0) {
				imgs.push(s);
			}
		});
		// Also check background-image in gallery
		document.querySelectorAll("[style*='background-image']").forEach(function(el) {
			var bg = el.style.backgroundImage || "";
			var urlM = bg.match(/url\(["']?(http[^"')]+)/);
			if (urlM && imgs.indexOf(urlM[1]) < 0) imgs.push(urlM[1]);
		});
		r.images = imgs.slice(0, 20);

		return JSON.stringify(r);
	})()`

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

	var result ListingDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result.Platform = "anjuke"
	result.URL = detailURL

	return &result, nil
}
