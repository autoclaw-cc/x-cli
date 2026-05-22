package property

import (
	"encoding/json"
	"fmt"
	"time"

	"rightmove-cli/browser"
)

type PropertyDetail struct {
	ID             string   `json:"id"`
	Platform       string   `json:"platform"`
	Title          string   `json:"title"`
	Address        string   `json:"address"`
	PricePCM       int      `json:"price_pcm"`
	PriceWeekly    int      `json:"price_weekly,omitempty"`
	Bedrooms       int      `json:"bedrooms"`
	Bathrooms      int      `json:"bathrooms,omitempty"`
	PropertyType   string   `json:"property_type,omitempty"`
	Agent          string   `json:"agent,omitempty"`
	AgentPhone     string   `json:"agent_phone,omitempty"`
	LetType        string   `json:"let_type,omitempty"`
	FurnishType    string   `json:"furnish_type,omitempty"`
	LetAvailable   string   `json:"let_available,omitempty"`
	Deposit        string   `json:"deposit,omitempty"`
	MinTenancy     string   `json:"min_tenancy,omitempty"`
	Features       []string `json:"features,omitempty"`
	Description    string   `json:"description,omitempty"`
	NearestStation string   `json:"nearest_station,omitempty"`
	Images         []string `json:"images,omitempty"`
	URL            string   `json:"url"`
}

func GetDetail(client *browser.Client, detailURL string) (*PropertyDetail, error) {
	if err := client.NavigateNewTab(detailURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(function() {
    var text = document.body ? document.body.innerText : "";
    if (text.length < 200) return JSON.stringify({error: "page_not_loaded"});
    if (/verify|robot|captcha/i.test(text) && text.length < 1000) {
        return JSON.stringify({error: "captcha_detected"});
    }

    // Extract property ID from URL
    var idMatch = window.location.pathname.match(/\/properties\/(\d+)/);
    var id = idMatch ? idMatch[1] : "";

    // --- Title and Address ---
    // The h1 typically contains something like "2 bed flat to rent" or the address
    var h1El = document.querySelector('h1');
    var title = h1El ? h1El.textContent.trim() : "";

    // Address is often in a separate element near the title
    var addressEl = document.querySelector('[class*=_address], [data-testid*=address], h1 + span, h1 + div, .property-header-bedroom-and-price + *');
    var address = addressEl ? addressEl.textContent.trim() : "";

    // Fallback: parse address from text
    if (!address) {
        // Look for a line with a postcode pattern (e.g., "SW1A 1AA" or "E1 6AN")
        var addrMatch = text.match(/([A-Z][A-Za-z0-9 ,]+,\s*[A-Z]{1,2}\d{1,2}[A-Z]?\s*\d[A-Z]{2})/);
        if (addrMatch) address = addrMatch[1].trim();
    }

    // --- Price ---
    var pricePCM = 0;
    var priceWeekly = 0;
    var pcmMatch = text.match(/£([\d,]+)\s*pcm/i);
    if (pcmMatch) pricePCM = parseInt(pcmMatch[1].replace(/,/g, "")) || 0;
    var pwMatch = text.match(/£([\d,]+)\s*pw/i);
    if (pwMatch) priceWeekly = parseInt(pwMatch[1].replace(/,/g, "")) || 0;
    // Fallback: just find a prominent price
    if (!pricePCM) {
        var priceAlt = text.match(/£([\d,]+)\s*per month/i);
        if (priceAlt) pricePCM = parseInt(priceAlt[1].replace(/,/g, "")) || 0;
    }

    // --- Bedrooms and Bathrooms ---
    var bedrooms = 0;
    var bathrooms = 0;
    var bedIdx = text.indexOf("BEDROOMS");
    if (bedIdx >= 0) {
        var afterBed = text.substring(bedIdx + 8, bedIdx + 20).replace(/[^\d]/g, "");
        if (afterBed) bedrooms = parseInt(afterBed.charAt(0)) || 0;
    }
    if (!bedrooms) {
        var bedMatch = text.match(/(\d+)\s*bed/i);
        if (bedMatch) bedrooms = parseInt(bedMatch[1]) || 0;
    }
    var bathIdxV = text.indexOf("BATHROOMS");
    if (bathIdxV >= 0) {
        var afterBath = text.substring(bathIdxV + 9, bathIdxV + 20).replace(/[^\d]/g, "");
        if (afterBath) bathrooms = parseInt(afterBath.charAt(0)) || 0;
    }
    if (!bathrooms) {
        var bathMatch = text.match(/(\d+)\s*bath/i);
        if (bathMatch) bathrooms = parseInt(bathMatch[1]) || 0;
    }

    // --- Property Type ---
    var propertyType = "";
    var typeMatch = text.match(/PROPERTY TYPE\s+([A-Za-z\s-]+)/i) || text.match(/Property type\s+([A-Za-z\s-]+)/i);
    if (typeMatch) {
        propertyType = typeMatch[1].trim().split("\n")[0].trim();
    }
    if (!propertyType) {
        // Try to extract from title: "2 bed flat", "3 bed house"
        var titleTypeMatch = title.match(/\d+\s*bed\s+([\w\s-]+?)(?:\s+to\s+rent|\s+in\s+|$)/i);
        if (titleTypeMatch) propertyType = titleTypeMatch[1].trim();
    }

    // --- Let Type, Furnish Type, Let Available ---
    var letType = "";
    var furnishType = "";
    var letAvailable = "";
    var deposit = "";
    var minTenancy = "";

    var letTypeMatch = text.match(/Let type\s*[:\n]?\s*([^\n]+)/i);
    if (letTypeMatch) letType = letTypeMatch[1].trim();
    var furnishMatch = text.match(/Furnish(?:ing)? type\s*[:\n]?\s*([^\n]+)/i);
    if (furnishMatch) furnishType = furnishMatch[1].trim();
    var letAvailMatch = text.match(/Let available[: ]*\s*\n?\s*([^\n]+)/i);
    if (letAvailMatch) letAvailable = letAvailMatch[1].trim();
    if (!letAvailable) {
        var availAlt = text.match(/Available\s+(Now|Immediately|\d{1,2}\/\d{1,2}\/\d{2,4})/i);
        if (availAlt) letAvailable = availAlt[1].trim();
    }
    var depositMatch = text.match(/Deposit\s*[:\n]?\s*£?([\d,]+)/i);
    if (depositMatch) deposit = "£" + depositMatch[1].trim();
    var tenancyMatch = text.match(/Min(?:imum)?\s*(?:term|tenancy)\s*[:\n]?\s*([^\n]+)/i);
    if (tenancyMatch) minTenancy = tenancyMatch[1].trim();

    // --- Agent Info ---
    var agent = "";
    var agentPhone = "";
    var agentEl = document.querySelector('[class*=_agentName], [class*=AgentName], [data-testid*=agent-name]');
    if (agentEl) agent = agentEl.textContent.trim();
    if (!agent) {
        var agentMatch = text.match(/(?:Marketed|Listed) by\s+([^\n]+)/i);
        if (agentMatch) agent = agentMatch[1].trim();
    }
    // Phone: look for UK phone number pattern
    var phoneEl = document.querySelector('[class*=_telephoneNumber], [class*=TelephoneNumber], [data-testid*=phone]');
    if (phoneEl) agentPhone = phoneEl.textContent.trim();
    if (!agentPhone) {
        var phoneMatch = text.match(/(0\d{2,4}\s*\d{3,4}\s*\d{3,4})/);
        if (phoneMatch) agentPhone = phoneMatch[1].trim();
    }

    // --- Key Features ---
    var features = [];
    var featureSection = text.match(/Key features\s*\n([\s\S]*?)(?:\n\n|\nLetting|About this|Description|Property description)/i);
    if (featureSection) {
        var featureLines = featureSection[1].split("\n").map(function(l){ return l.trim(); }).filter(function(l){ return l && l.length > 2; });
        features = featureLines;
    }
    // Fallback: try DOM-based feature extraction
    if (features.length === 0) {
        document.querySelectorAll('[class*=KeyFeature] li, [class*=keyFeature] li, [class*=_feature] li, ul[class*=feature] li').forEach(function(li) {
            var t = li.textContent.trim();
            if (t) features.push(t);
        });
    }

    // --- Description ---
    var description = "";
    var descSection = text.match(/(?:About this property|Property description|Description)\s*\n([\s\S]*?)(?:\n\n(?:Key features|Letting information|Property info|Nearest station|Floor plan)|$)/i);
    if (descSection) {
        description = descSection[1].trim();
        // Limit description length
        if (description.length > 2000) description = description.substring(0, 2000) + "...";
    }

    // --- Nearest Station ---
    var nearestStation = "";
    var stationMatch = text.match(/Nearest station[s]?\s*\n?\s*([^\n]+)/i);
    if (stationMatch) {
        nearestStation = stationMatch[1].trim();
        // Try to include distance
        var stationLines = text.split(stationMatch[0]);
        if (stationLines.length > 1) {
            var nextLine = stationLines[1].split("\n").filter(function(l){ return l.trim(); })[0];
            if (nextLine && /mile|km|metre|yard|walk|min/i.test(nextLine)) {
                nearestStation += " - " + nextLine.trim();
            }
        }
    }

    // --- Images ---
    var images = [];
    document.querySelectorAll('[class*=Gallery] img, [class*=gallery] img, [class*=Slideshow] img, [class*=carousel] img, picture img').forEach(function(img) {
        var src = img.src || img.dataset.src || img.getAttribute("srcset") || "";
        // Get the first URL from srcset if present
        if (src.includes(",")) src = src.split(",")[0].trim().split(" ")[0];
        if (src && !src.includes("svg") && !src.includes("data:") && src.startsWith("http")) {
            // Deduplicate
            if (images.indexOf(src) === -1) images.push(src);
        }
    });
    // Fallback: look for media URLs in the page
    if (images.length === 0) {
        document.querySelectorAll('img[src*="media.rightmove"]').forEach(function(img) {
            var src = img.src;
            if (src && images.indexOf(src) === -1) images.push(src);
        });
    }

    return JSON.stringify({
        id: id,
        platform: "rightmove",
        title: title,
        address: address,
        price_pcm: pricePCM,
        price_weekly: priceWeekly,
        bedrooms: bedrooms,
        bathrooms: bathrooms,
        property_type: propertyType,
        agent: agent,
        agent_phone: agentPhone,
        let_type: letType,
        furnish_type: furnishType,
        let_available: letAvailable,
        deposit: deposit,
        min_tenancy: minTenancy,
        features: features,
        description: description,
        nearest_station: nearestStation,
        images: images,
        url: %q
    });
})()`, detailURL)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	// Check for page-level errors
	var check struct {
		Error string `json:"error"`
	}
	json.Unmarshal(raw, &check)
	if check.Error == "captcha_detected" {
		return nil, fmt.Errorf("Rightmove CAPTCHA detected. Please complete the verification in Chrome, then retry.")
	}
	if check.Error == "page_not_loaded" {
		return nil, fmt.Errorf("page did not load properly; body text too short")
	}
	if check.Error != "" {
		return nil, fmt.Errorf("page issue: %s", check.Error)
	}

	var detail PropertyDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &detail, nil
}
