package house

import (
	"58-cli/browser"
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
		if (/验证码|安全验证|人机验证|验证/.test(text) && text.length < 500) {
			return JSON.stringify({error: "captcha_required"});
		}

		var r = {};
		var m;

		// title: first meaningful line after breadcrumb
		var lines = text.split("\n").filter(function(l){ return l.trim().length > 0; });
		var breadcrumbIdx = -1;
		for (var i = 0; i < lines.length; i++) {
			if (lines[i].indexOf("整租房") >= 0 && lines[i].indexOf(">") >= 0) {
				breadcrumbIdx = i;
				break;
			}
		}
		r.title = breadcrumbIdx >= 0 && breadcrumbIdx + 1 < lines.length ? lines[breadcrumbIdx + 1].trim() : "";

		// price
		m = text.match(/([\d,]+)\s*元\/月/);
		r.rent_monthly = m ? parseInt(m[1].replace(/,/g, "")) : 0;

		// deposit model
		m = text.match(/押[一二三]付[一二三]|半年付|年付|押一付一/);
		r.deposit_model = m ? m[0] : "";

		// rent type
		m = text.match(/租赁方式：(\S+)/);
		r.rent_type = m ? m[1] : "";

		// layout, area, decoration
		m = text.match(/房屋类型：([^\n]+)/);
		var layoutLine = m ? m[1].trim() : "";
		r.layout = "";
		r.area_sqm = 0;
		r.decoration = "";
		if (layoutLine) {
			var lm = layoutLine.match(/(\d+室\d*厅?\d*卫?)/);
			r.layout = lm ? lm[1] : "";
			var am = layoutLine.match(/([\d.]+)\s*平/);
			r.area_sqm = am ? parseFloat(am[1]) : 0;
			if (/精装/.test(layoutLine)) r.decoration = "精装";
			else if (/简装/.test(layoutLine)) r.decoration = "简装";
			else if (/毛坯/.test(layoutLine)) r.decoration = "毛坯";
			else if (/豪装/.test(layoutLine)) r.decoration = "豪装";
		}

		// floor and orientation
		m = text.match(/朝向楼层：\s*([^\n]+)/);
		r.floor = "";
		r.orientation = "";
		if (m) {
			var floorLine = m[1].trim();
			var om = floorLine.match(/^(\S+)\s/);
			r.orientation = om ? om[1] : "";
			var fm = floorLine.match(/(低|中|高)层\s*\/\s*(\d+)层/);
			r.floor = fm ? fm[1] + "层/" + fm[2] + "层" : floorLine;
		}

		// compound
		m = text.match(/所在小区：\s*([^(（\n]+)/);
		r.compound = m ? m[1].trim() : "";

		// district + subway
		m = text.match(/所属区域：([^\n]+)/);
		r.district = m ? m[1].trim().replace(/\s+/g, " ") : "";

		// subway info from district line
		r.nearby = {subway: []};
		var subwayMatch = r.district.match(/距离?地铁(\d+号线)([^\d]+?)站?(\d+)米/);
		if (subwayMatch) {
			r.nearby.subway.push({
				line: subwayMatch[1],
				name: subwayMatch[2].replace(/站$/, ""),
				distance: subwayMatch[3] + "m"
			});
		}

		// address
		m = text.match(/详细地址：\s*([^\n]+)/);
		r.address = m ? m[1].trim().replace(/\s*附近高薪工作.*/, "").replace(/\s*查看地图.*/, "") : "";

		// agent — match (经纪人), (联系人), (品牌公寓), (个人) etc.
		r.poster = {name: "", type: "unknown", company: "", listing_count: 0, verified: false};
		m = text.match(/(\S+)\(经纪人\)/);
		if (m) {
			r.poster.name = m[1];
			r.poster.type = "agent";
			var cm = text.match(/\(经纪人\)\s*\n\s*([^\n]+)/);
			r.poster.company = cm ? cm[1].trim() : "";
		} else {
			m = text.match(/(\S+)\(联系人\)/);
			if (m) {
				r.poster.name = m[1];
				r.poster.type = "agent";
				var cm2 = text.match(/\(联系人\)\s*\n\s*([^\n]+)/);
				r.poster.company = cm2 ? cm2[1].trim() : "";
			} else {
				m = text.match(/(\S+)\(品牌公寓\)/);
				if (m) {
					r.poster.name = m[1];
					r.poster.type = "brand";
					var cm3 = text.match(/\(品牌公寓\)\s*\n\s*([^\n]+)/);
					r.poster.company = cm3 ? cm3[1].trim() : "";
				} else {
					m = text.match(/(\S+)\(个人\)/);
					if (m) {
						r.poster.name = m[1];
						r.poster.type = "owner";
					}
				}
			}
		}

		// verified badge
		if (/实名认证|已认证/.test(text)) r.poster.verified = true;
		if (/营业执照/.test(text)) r.poster.verified = true;

		// listing count for compound
		m = text.match(/在租\s*(\d+)\s*套/) || text.match(/(\d+)\s*\n\s*在租房源/);
		r.poster.listing_count = m ? parseInt(m[1]) : 0;

		// id
		m = text.match(/房源编码：\s*(\d+)/);
		r.id = m ? m[1] : "";

		// highlights / features
		m = text.match(/房屋亮点\s*\n([^\n]+)/);
		r.features = m ? m[1].trim().split(/\s+/) : [];

		// description
		m = text.match(/房源描述\s*\n([\s\S]*?)(?=\n小区:|\n周边|\n猜你喜欢|$)/);
		r.description = m ? m[1].trim().substring(0, 1000) : "";

		// built year and biz circle
		m = text.match(/建筑年代：(\d+)/);
		r.built_year = m ? m[1] : "";
		m = text.match(/所属商圈：([^\n]+)/);
		r.biz_circle = m ? m[1].trim() : "";

		// images
		var imgs = [];
		document.querySelectorAll("img").forEach(function(el) {
			var s = el.src || el.getAttribute("data-src") || "";
			if (s.indexOf("58cdn") >= 0 && s.indexOf("lazy_pic") < 0 && imgs.indexOf(s) < 0) {
				imgs.push(s);
			}
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
		return nil, fmt.Errorf("58.com 安全验证触发，请在 Chrome 中完成验证后重试")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s %s", check.Error, check.Message)
	}

	var result ListingDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result.Platform = "58"
	result.URL = detailURL

	return &result, nil
}
