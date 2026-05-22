package property

import (
	"apartments-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type Unit struct {
	Layout    string `json:"layout"`
	Price     string `json:"price"`
	SqFt      string `json:"sqft,omitempty"`
	Available string `json:"available,omitempty"`
}

type PropertyDetail struct {
	ID          string   `json:"id"`
	Platform    string   `json:"platform"`
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	RentRange   string   `json:"rent_range,omitempty"`
	BedRange    string   `json:"bed_range,omitempty"`
	BathRange   string   `json:"bath_range,omitempty"`
	SqftRange   string   `json:"sqft_range,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
	Amenities   []string `json:"amenities,omitempty"`
	Description string   `json:"description,omitempty"`
	Units       []Unit   `json:"units,omitempty"`
	Images      []string `json:"images,omitempty"`
	URL         string   `json:"url"`
}

func GetDetail(client *browser.Client, detailURL string) (*PropertyDetail, error) {
	if err := client.NavigateNewTab(detailURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}
	time.Sleep(5 * time.Second)

	js := `(function(){
		var text = document.body ? document.body.innerText : "";
		if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
		if (/verify|robot|captcha/i.test(text) && text.length < 1000) {
			return JSON.stringify({error: "captcha_detected"});
		}

		var r = {};

		// ID from URL
		var pathParts = location.pathname.split("/").filter(function(p){return p});
		r.id = pathParts.length > 0 ? pathParts[pathParts.length - 1] : "";

		// Name
		var h1 = document.querySelector("h1");
		r.name = h1 ? h1.innerText.trim() : "";

		// Address
		var addrEl = document.querySelector(".delivery-address") || document.querySelector(".propertyAddress");
		r.address = addrEl ? addrEl.innerText.trim() : "";

		// Rent/Bed/Bath/Sqft ranges
		var rentEl = document.querySelector(".rentInfoDetail");
		r.rent_range = rentEl ? rentEl.innerText.trim().split("\n")[0] : "";

		var bedEl = document.querySelector(".bedInfoDetail");
		r.bed_range = bedEl ? bedEl.innerText.trim().split("\n")[0] : "";

		var bathEl = document.querySelector(".bathInfoDetail");
		r.bath_range = bathEl ? bathEl.innerText.trim().split("\n")[0] : "";

		var sqftEl = document.querySelector(".sqftInfoDetail");
		r.sqft_range = sqftEl ? sqftEl.innerText.trim().split("\n")[0] : "";

		// Phone
		var phoneEl = document.querySelector(".phoneNumber");
		r.phone = phoneEl ? phoneEl.innerText.trim() : "";

		// Highlights (unique amenities)
		r.highlights = [];
		document.querySelectorAll(".specInfo, .uniqueAmenity").forEach(function(el) {
			var t = el.innerText.trim();
			if (t && r.highlights.indexOf(t) < 0) r.highlights.push(t);
		});

		// Amenities (general)
		r.amenities = [];
		document.querySelectorAll(".amenityCard, .amenity-group").forEach(function(el) {
			var t = el.innerText.trim().split("\n")[0];
			if (t && r.amenities.indexOf(t) < 0) r.amenities.push(t);
		});

		// Description
		var descEl = document.querySelector(".descriptionSection p, .propertyDescription");
		r.description = descEl ? descEl.innerText.trim() : "";
		if (!r.description) {
			var aboutMatch = text.match(/About [^\n]+\n\n([\s\S]*?)(?=\n\nRead More|\n\nRent Specials|\n\nHighlights|\n\nPricing|$)/);
			if (aboutMatch) r.description = aboutMatch[1].trim();
		}
		if (r.description.length > 1500) r.description = r.description.substring(0, 1500) + "...";

		// Units from pricing grid
		r.units = [];
		document.querySelectorAll(".pricingGridItem").forEach(function(pg) {
			var pgText = pg.innerText.trim();
			var layoutMatch = pgText.match(/^([^\n]+)/);
			var layout = layoutMatch ? layoutMatch[1] : "";

			var priceMatch = pgText.match(/\$([\d,]+)\s*[–-]\s*\$([\d,]+)/) || pgText.match(/\$([\d,]+)/);
			var price = priceMatch ? priceMatch[0] : "";

			var unitRows = pg.querySelectorAll("tr, .unitContainer");
			if (unitRows.length > 0) {
				unitRows.forEach(function(row) {
					var rowText = row.innerText.trim();
					var unitPrice = "";
					var unitSqft = "";
					var unitAvail = "";
					var pm = rowText.match(/\$([\d,]+)/);
					if (pm) unitPrice = pm[0];
					var sqftEl = row.querySelector(".sqftColumn span, .sqftColumn");
					if (sqftEl) {
						var sv = sqftEl.innerText.trim().replace(/[^\d]/g, "");
						if (sv && sv !== "0") unitSqft = sv;
					}
					if (/Now|Available|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec|\d{1,2}\/\d{1,2}/i.test(rowText)) {
						var am = rowText.match(/(Now|Available Now|\d{1,2}\/\d{1,2}(?:\/\d{2,4})?|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2})/i);
						if (am) unitAvail = am[1];
					}
					if (unitPrice) {
						r.units.push({layout: layout, price: unitPrice, sqft: unitSqft, available: unitAvail});
					}
				});
			} else if (price) {
				r.units.push({layout: layout, price: price, sqft: "", available: ""});
			}
		});

		// Images
		r.images = [];
		var seen = {};
		document.querySelectorAll("img").forEach(function(el) {
			var s = el.src || el.getAttribute("data-src") || "";
			if (s && s.indexOf("http") === 0 && s.indexOf("svg") < 0 && s.indexOf("data:") < 0 &&
				s.indexOf("logo") < 0 && s.indexOf("icon") < 0 && s.indexOf("avatar") < 0 &&
				!seen[s] && (s.indexOf("cloudfront") >= 0 || s.indexOf("apartments.com") >= 0 || s.indexOf("apartmenthomeliving") >= 0)) {
				seen[s] = true;
				r.images.push(s);
			}
		});
		r.images = r.images.slice(0, 20);

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
	if check.Error == "captcha_detected" {
		return nil, fmt.Errorf("Apartments.com CAPTCHA detected. Please complete verification in Chrome.")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s %s", check.Error, check.Message)
	}

	var result PropertyDetail
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result.Platform = "apartments.com"
	result.URL = detailURL

	return &result, nil
}
