package property

import (
	"apartments-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type PriceOption struct {
	Beds  string `json:"beds"`
	Price string `json:"price"`
}

type Listing struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Address   string        `json:"address"`
	Pricing   []PriceOption `json:"pricing"`
	Phone     string        `json:"phone,omitempty"`
	Amenities string        `json:"amenities,omitempty"`
	URL       string        `json:"url"`
	Image     string        `json:"image,omitempty"`
}

type SearchResult struct {
	Properties []Listing `json:"properties"`
	Location   string    `json:"location"`
	TotalFound string    `json:"total_found"`
	Page       int       `json:"page"`
}

type SearchParams struct {
	Location string
	MinBeds  int
	MaxBeds  int
	MinPrice int
	MaxPrice int
	Limit    int
	Page     int
}

func Search(client *browser.Client, params SearchParams) (*SearchResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	searchURL := fmt.Sprintf("https://www.apartments.com/%s/", params.Location)

	filterParts := ""
	if params.MinBeds > 0 && params.MaxBeds > 0 && params.MinBeds == params.MaxBeds {
		filterParts += fmt.Sprintf("%d-bedrooms-", params.MinBeds)
	} else if params.MinBeds > 0 {
		filterParts += fmt.Sprintf("min-%d-bedrooms-", params.MinBeds)
	} else if params.MaxBeds > 0 {
		filterParts += fmt.Sprintf("max-%d-bedrooms-", params.MaxBeds)
	}

	if params.MinPrice > 0 && params.MaxPrice > 0 {
		filterParts += fmt.Sprintf("%d-to-%d/", params.MinPrice, params.MaxPrice)
	} else if params.MinPrice > 0 {
		filterParts += fmt.Sprintf("over-%d/", params.MinPrice)
	} else if params.MaxPrice > 0 {
		filterParts += fmt.Sprintf("under-%d/", params.MaxPrice)
	}

	if filterParts != "" {
		if filterParts[len(filterParts)-1] != '/' {
			filterParts += "/"
		}
		searchURL += filterParts
	}

	if params.Page > 1 {
		searchURL += fmt.Sprintf("%d/", params.Page)
	}

	if err := client.NavigateNewTab(searchURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(function(){
		var text = document.body ? document.body.innerText : "";
		if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
		if (/verify|robot|captcha/i.test(text) && text.length < 1000) {
			return JSON.stringify({error: "captcha_detected"});
		}

		var cards = document.querySelectorAll("[data-listingid]");
		if (!cards.length) return JSON.stringify({error: "no_listings", message: "no [data-listingid] found"});

		var properties = [];
		var limit = %d;

		cards.forEach(function(c) {
			if (properties.length >= limit) return;

			var id = c.getAttribute("data-listingid") || "";
			var titleEl = c.querySelector(".js-placardTitle");
			var name = titleEl ? titleEl.innerText.trim() : "";
			var addrEl = c.querySelector(".property-address");
			var address = addrEl ? addrEl.innerText.trim() : "";

			var pricing = [];
			c.querySelectorAll(".bedRentBox").forEach(function(br) {
				var lines = br.innerText.trim().split("\n").map(function(l){return l.trim()}).filter(function(l){return l});
				if (lines.length >= 2) {
					pricing.push({beds: lines[0], price: lines[1]});
				}
			});

			var phoneEl = c.querySelector(".js-phone");
			var phone = phoneEl ? phoneEl.innerText.trim() : "";

			var amenEl = c.querySelector(".property-amenities");
			var amenities = amenEl ? amenEl.innerText.trim() : "";

			var linkEl = c.querySelector("a");
			var url = linkEl ? linkEl.href.split("?")[0] : "";

			var imgEl = c.querySelector(".carousel-item img, img.loaded");
			var image = "";
			if (imgEl) {
				image = imgEl.src || imgEl.getAttribute("data-src") || "";
				if (image.indexOf("data:") === 0) image = "";
			}

			if (id) {
				properties.push({
					id: id,
					name: name,
					address: address,
					pricing: pricing,
					phone: phone,
					amenities: amenities,
					url: url,
					image: image
				});
			}
		});

		var totalText = "";
		var titleMatch = document.title.match(/([\d,]+)\s*Rental/i);
		if (titleMatch) {
			totalText = titleMatch[1];
		} else {
			var resultEl = document.querySelector(".searchResults .resultCount, .resultCount, #placardHeader h1");
			totalText = resultEl ? resultEl.innerText.trim() : String(properties.length);
			var numMatch = totalText.match(/([\d,]+)/);
			if (numMatch) totalText = numMatch[1];
		}

		return JSON.stringify({
			properties: properties,
			total_found: totalText
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
	if check.Error == "captcha_detected" {
		return nil, fmt.Errorf("Apartments.com CAPTCHA detected. Please complete verification in Chrome.")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s %s", check.Error, check.Message)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result.Location = params.Location
	result.Page = params.Page

	return &result, nil
}
