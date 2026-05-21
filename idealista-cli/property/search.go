package property

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"idealista-cli/browser"
)

type Property struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	PriceMonthly int      `json:"price_monthly"`
	Rooms        int      `json:"rooms"`
	Bathrooms    int      `json:"bathrooms,omitempty"`
	AreaSqm      float64  `json:"area_sqm"`
	Floor        string   `json:"floor,omitempty"`
	HasLift      bool     `json:"has_lift,omitempty"`
	Agent        string   `json:"agent,omitempty"`
	URL          string   `json:"url"`
	Images       []string `json:"images,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type SearchResult struct {
	Properties []Property `json:"properties"`
	Country    string     `json:"country"`
	City       string     `json:"city"`
	TotalFound string     `json:"total_found"`
	Page       int        `json:"page"`
}

type SearchParams struct {
	Country  string
	City     string
	MinPrice int
	MaxPrice int
	MinRooms int
	MaxRooms int
	Limit    int
	Page     int
}

var CountryDomains = map[string]string{
	"spain":    "www.idealista.com",
	"italy":    "www.idealista.it",
	"portugal": "www.idealista.pt",
}

func Search(client *browser.Client, params SearchParams) (*SearchResult, error) {
	domain, ok := CountryDomains[params.Country]
	if !ok {
		return nil, fmt.Errorf("unsupported country: %s", params.Country)
	}

	var searchURL string
	switch params.Country {
	case "italy":
		searchURL = fmt.Sprintf("https://%s/affitto-case/%s/", domain, params.City)
	case "portugal":
		searchURL = fmt.Sprintf("https://%s/arrendar-casas/%s/", domain, params.City)
	default: // spain
		searchURL = fmt.Sprintf("https://%s/alquiler-viviendas/%s/", domain, params.City)
	}

	// Add query-string filters
	queryParts := []string{}
	if params.MinPrice > 0 {
		queryParts = append(queryParts, fmt.Sprintf("minPrice=%d", params.MinPrice))
	}
	if params.MaxPrice > 0 {
		queryParts = append(queryParts, fmt.Sprintf("maxPrice=%d", params.MaxPrice))
	}
	if params.MinRooms > 0 {
		queryParts = append(queryParts, fmt.Sprintf("minSize=%d", params.MinRooms))
	}

	// Add page via path segment (pagina-N.htm before query string)
	if params.Page > 1 {
		searchURL += fmt.Sprintf("pagina-%d.htm", params.Page)
	}

	if len(queryParts) > 0 {
		searchURL += "?" + strings.Join(queryParts, "&")
	}

	if err := client.NavigateNewTab(searchURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(5 * time.Second)

	js := fmt.Sprintf(`(() => {
		// --- CAPTCHA detection ---
		const curURL = window.location.href.toLowerCase();
		const bodyText = document.body ? document.body.innerText : '';
		if (curURL.includes('captcha') || curURL.includes('robot') ||
		    (bodyText.length < 500 && (bodyText.toLowerCase().includes('captcha') || bodyText.toLowerCase().includes('robot')))) {
			return JSON.stringify({
				properties: [],
				country: '%s',
				city: '%s',
				total_found: 'CAPTCHA_DETECTED',
				page: %d
			});
		}

		const limit = %d;
		// Primary selector: .item.extended-item; fallback: article.item
		let items = document.querySelectorAll('.item.extended-item');
		if (items.length === 0) items = document.querySelectorAll('article.item');
		if (items.length === 0) items = document.querySelectorAll('.item-multimedia-container').length > 0
			? document.querySelectorAll('.item') : [];

		const properties = [];
		const count = Math.min(items.length, limit);

		for (let i = 0; i < count; i++) {
			const el = items[i];

			// --- Link & ID ---
			const linkEl = el.querySelector('a.item-link') || el.querySelector('a[href*="/inmueble/"]') || el.querySelector('a[href*="/immobile/"]') || el.querySelector('a[href*="/imovel/"]');
			const href = linkEl ? linkEl.getAttribute('href') : '';
			const fullURL = href ? new URL(href, window.location.origin).href : '';
			const idMatch = href ? href.match(/\/(\d+)\/?/) : null;
			const id = idMatch ? idMatch[1] : (el.getAttribute('data-adid') || el.getAttribute('data-element-id') || '');

			// --- Title / Address ---
			const titleEl = linkEl || el.querySelector('.item-link');
			const title = titleEl ? titleEl.textContent.trim() : '';

			// --- Price ---
			const priceEl = el.querySelector('.item-price, .price-row');
			let priceText = priceEl ? priceEl.textContent : '';
			// Extract the first number (current price), ignoring old/crossed-out price
			const priceMatch = priceText.match(/([\d.,]+)\s*€/);
			const price = priceMatch ? parseInt(priceMatch[1].replace(/[.,]/g, ''), 10) || 0 : 0;

			// --- Features line: "N bed. N m2 Xth floor exterior with lift" ---
			let rooms = 0, area = 0, floor = '', hasLift = false;
			const tags = [];

			// Look for the detail items / feature spans
			const detailEls = el.querySelectorAll('.item-detail span, .item-detail small, .item-detail');
			const allDetailText = [];
			detailEls.forEach(d => { allDetailText.push(d.textContent.trim()); });
			const combinedDetail = allDetailText.join(' ');

			// Rooms: "N bed" or "N hab"
			const roomMatch = combinedDetail.match(/(\d+)\s*(?:bed|hab|local|room)/i);
			if (roomMatch) rooms = parseInt(roomMatch[1], 10) || 0;

			// Area: "N m2" or "N m²"
			const areaMatch = combinedDetail.match(/([\d.,]+)\s*m[²2]/i);
			if (areaMatch) area = parseFloat(areaMatch[1].replace(',', '.')) || 0;

			// Floor: e.g. "Ground floor", "5th floor", "1st floor", "planta 3", "piano 2"
			const floorMatch = combinedDetail.match(/((?:ground|basement|\d+(?:st|nd|rd|th)?)\s*floor|planta\s*\w+|piano\s*\w+|andar\s*\w+)/i);
			if (floorMatch) floor = floorMatch[0].trim();

			// Lift
			if (/with lift|con ascensor|con ascensore|com elevador/i.test(combinedDetail)) hasLift = true;

			// Tags: exterior/interior
			if (/exterior/i.test(combinedDetail)) tags.push('exterior');
			if (/interior/i.test(combinedDetail)) tags.push('interior');

			// --- Agent ---
			const agentEl = el.querySelector('.logo-branding img, .professional-logo img');
			const agent = agentEl ? (agentEl.getAttribute('alt') || agentEl.getAttribute('title') || '') : '';

			properties.push({
				id: id,
				title: title,
				address: title,
				price_monthly: price,
				rooms: rooms,
				bathrooms: 0,
				area_sqm: area,
				floor: floor,
				has_lift: hasLift,
				agent: agent,
				url: fullURL,
				images: [],
				tags: tags
			});
		}

		// --- Total found ---
		const totalEl = document.querySelector('.listing-title .total-results')
			|| document.querySelector('.breadcrumb-subitems .total-results')
			|| document.querySelector('#h1-container h1')
			|| document.querySelector('h1');
		let totalText = totalEl ? totalEl.textContent.trim() : '';
		// Try to extract just the number from e.g. "12,345 results"
		const totalMatch = totalText.match(/([\d.,]+)\s*(?:results|properties|pisos|inmuebles|case|imóveis|anúncios)/i);
		if (totalMatch) totalText = totalMatch[1].replace(/\./g, '').replace(/,/g, '');

		return JSON.stringify({
			properties: properties,
			country: '%s',
			city: '%s',
			total_found: totalText,
			page: %d
		});
	})()`, params.Country, params.City, params.Page, params.Limit, params.Country, params.City, params.Page)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &result, nil
}
