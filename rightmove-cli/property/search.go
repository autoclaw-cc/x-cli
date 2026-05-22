package property

import (
	"encoding/json"
	"fmt"
	"time"

	"rightmove-cli/browser"
)

type Property struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	PricePCM     int      `json:"price_pcm"`
	PriceWeekly  int      `json:"price_weekly,omitempty"`
	Bedrooms     int      `json:"bedrooms"`
	Bathrooms    int      `json:"bathrooms,omitempty"`
	PropertyType string   `json:"property_type"`
	Agent        string   `json:"agent"`
	LetType      string   `json:"let_type,omitempty"`
	FurnishType  string   `json:"furnish_type,omitempty"`
	URL          string   `json:"url"`
	Images       []string `json:"images,omitempty"`
	LetAvailable string   `json:"let_available,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type SearchResult struct {
	Properties []Property `json:"properties"`
	Location   string     `json:"location"`
	TotalFound string     `json:"total_found"`
	Page       int        `json:"page"`
}

type SearchParams struct {
	Location string
	MinPrice int
	MaxPrice int
	MinBeds  int
	MaxBeds  int
	Radius   float64
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

	searchURL := fmt.Sprintf(
		"https://www.rightmove.co.uk/property-to-rent/%s.html",
		params.Location,
	)
	queryParts := []string{}
	if params.MinPrice > 0 {
		queryParts = append(queryParts, fmt.Sprintf("minPrice=%d", params.MinPrice))
	}
	if params.MaxPrice > 0 {
		queryParts = append(queryParts, fmt.Sprintf("maxPrice=%d", params.MaxPrice))
	}
	if params.MinBeds > 0 {
		queryParts = append(queryParts, fmt.Sprintf("minBedrooms=%d", params.MinBeds))
	}
	if params.MaxBeds > 0 {
		queryParts = append(queryParts, fmt.Sprintf("maxBedrooms=%d", params.MaxBeds))
	}
	if params.Radius > 0 {
		queryParts = append(queryParts, fmt.Sprintf("radius=%.1f", params.Radius))
	}
	if params.Page > 1 {
		queryParts = append(queryParts, fmt.Sprintf("index=%d", (params.Page-1)*24))
	}
	if len(queryParts) > 0 {
		searchURL += "?"
		for i, p := range queryParts {
			if i > 0 {
				searchURL += "&"
			}
			searchURL += p
		}
	}

	if err := client.NavigateNewTab(searchURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(function() {
    var text = document.body ? document.body.innerText : "";
    if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
    if (/verify|robot|captcha/i.test(text) && text.length < 1000) {
        return JSON.stringify({error: "captcha_detected"});
    }

    var limit = %d;
    var cards = document.querySelectorAll('[class*=PropertyCard_propertyCardContainer], [class*=propertyCard-wrapper], .l-searchResult');
    if (cards.length === 0) {
        // Fallback: try text-based parsing
        return JSON.stringify({error: "no_cards_found", hint: "Cards selector did not match. Page text length: " + text.length});
    }

    var properties = [];
    var seenIds = {};

    for (var i = 0; i < cards.length; i++) {
        if (properties.length >= limit) break;
        var card = cards[i];
        var cardText = card.innerText || "";
        var lines = cardText.split("\n").map(function(l){ return l.trim(); }).filter(function(l){ return l; });

        var anchor = card.querySelector('a[href*="/properties/"]');
        var href = anchor ? anchor.href : "";
        var idMatch = href.match(/\/properties\/(\d+)/);
        var id = idMatch ? idMatch[1] : "";
        var propURL = href ? href.split("#")[0] : "";

        if (id && seenIds[id]) continue;
        if (id) seenIds[id] = true;

        // Parse price PCM
        var pricePCM = 0;
        var priceWeekly = 0;
        var pcmMatch = cardText.match(/£([\d,]+)\s*pcm/i);
        if (pcmMatch) pricePCM = parseInt(pcmMatch[1].replace(/,/g, "")) || 0;
        var pwMatch = cardText.match(/£([\d,]+)\s*pw/i);
        if (pwMatch) priceWeekly = parseInt(pwMatch[1].replace(/,/g, "")) || 0;

        // Parse address: look for line with comma that looks like an address
        var address = "";
        var propertyType = "";
        var bedrooms = 0;
        var bathrooms = 0;
        var agent = "";
        var description = "";

        for (var j = 0; j < lines.length; j++) {
            var line = lines[j];

            // Address: typically contains comma and postcode-like pattern or area name
            // It usually comes after the price lines
            if (!address && /,/.test(line) && !/pcm|pw|Added|by |£/.test(line)) {
                address = line;
            }

            // Property type: standalone words like Apartment, Flat, House, Studio, etc.
            if (!propertyType && /^(Apartment|Flat|House|Detached|Semi-Detached|Terraced|Bungalow|Maisonette|Studio|Penthouse|Duplex|Room|Cottage|Villa|Town House|End of Terrace|Mews)$/i.test(line)) {
                propertyType = line;
            }

            // Bedrooms/bathrooms: standalone numbers following property type
            if (propertyType && !bedrooms && /^\d+$/.test(line)) {
                bedrooms = parseInt(line);
            } else if (propertyType && bedrooms && !bathrooms && /^\d+$/.test(line)) {
                bathrooms = parseInt(line);
            }

            // Agent info: "Added on DD/MM/YYYY by ..."
            var agentMatch = line.match(/Added on .+ by (.+)/);
            if (agentMatch) {
                agent = agentMatch[1].trim();
            }

            // Description: longer text line that is not price/agent/address
            if (!description && line.length > 30 && !/£|pcm|pw|Added on|^\d+$/.test(line) && line !== address && line !== propertyType) {
                description = line;
            }
        }

        // Build title from address + property type
        var title = "";
        if (propertyType && bedrooms) {
            title = bedrooms + " bed " + propertyType;
            if (address) title += " in " + address.split(",")[0].trim();
        } else if (address) {
            title = address;
        }

        properties.push({
            id: id,
            title: title,
            address: address,
            price_pcm: pricePCM,
            price_weekly: priceWeekly,
            bedrooms: bedrooms,
            bathrooms: bathrooms,
            property_type: propertyType,
            agent: agent,
            url: propURL,
            tags: description ? [description] : []
        });
    }

    var totalText = "";
    var headingMatch = text.match(/([\d,]+)\s*(rental |)propert/i);
    if (headingMatch) {
        totalText = headingMatch[1];
    } else {
        var totalEl = document.querySelector('[class*=searchHeader] [class*=count], .searchHeader-resultCount');
        totalText = totalEl ? totalEl.textContent.trim() : String(properties.length);
    }

    return JSON.stringify({
        properties: properties,
        location: %q,
        total_found: totalText,
        page: %d
    });
})()`, params.Limit, params.Location, params.Page)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	// Check for page-level errors
	var check struct {
		Error string `json:"error"`
		Hint  string `json:"hint"`
	}
	json.Unmarshal(raw, &check)
	if check.Error == "captcha_detected" {
		return nil, fmt.Errorf("Rightmove CAPTCHA detected. Please complete the verification in Chrome, then retry.")
	}
	if check.Error == "page_not_loaded" {
		return nil, fmt.Errorf("page did not load properly; body text too short")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s (%s)", check.Error, check.Hint)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &result, nil
}
