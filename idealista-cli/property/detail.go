package property

import (
	"encoding/json"
	"fmt"
	"time"

	"idealista-cli/browser"
)

type PropertyDetail struct {
	ID           string   `json:"id"`
	Platform     string   `json:"platform"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	PriceMonthly int      `json:"price_monthly"`
	Rooms        int      `json:"rooms"`
	Bathrooms    int      `json:"bathrooms,omitempty"`
	AreaSqm      float64  `json:"area_sqm"`
	Floor        string   `json:"floor,omitempty"`
	HasLift      bool     `json:"has_lift,omitempty"`
	Furnished    string   `json:"furnished,omitempty"`
	Deposit      string   `json:"deposit,omitempty"`
	Agent        string   `json:"agent,omitempty"`
	AgentPhone   string   `json:"agent_phone,omitempty"`
	Features     []string `json:"features,omitempty"`
	Description  string   `json:"description,omitempty"`
	EnergyRating string   `json:"energy_rating,omitempty"`
	Images       []string `json:"images,omitempty"`
	URL          string   `json:"url"`
}

func Detail(client *browser.Client, url string) (*PropertyDetail, error) {
	if err := client.NavigateNewTab(url); err != nil {
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
				id: '', platform: 'idealista', title: 'CAPTCHA_DETECTED',
				address: '', price_monthly: 0, rooms: 0, bathrooms: 0,
				area_sqm: 0, floor: '', has_lift: false, furnished: '',
				deposit: '', agent: '', agent_phone: '', features: [],
				description: 'Page blocked by CAPTCHA. Please solve it in the browser and retry.',
				energy_rating: '', images: [], url: '%s'
			});
		}

		// --- ID from URL path ---
		const idMatch = window.location.pathname.match(/\/(\d+)\/?/);
		const id = idMatch ? idMatch[1] : '';

		// --- Title ---
		const titleEl = document.querySelector('.main-info__title-main')
			|| document.querySelector('.detail-info-title h1')
			|| document.querySelector('h1.h3-simulated')
			|| document.querySelector('h1');
		const title = titleEl ? titleEl.textContent.trim() : '';

		// --- Address ---
		const addrEl = document.querySelector('.main-info__title-minor')
			|| document.querySelector('.header-map-list')
			|| document.querySelector('.main-info .subtitle');
		const address = addrEl ? addrEl.textContent.trim() : '';

		// --- Price ---
		const priceEl = document.querySelector('.info-data-price .txt-big')
			|| document.querySelector('.price .txt-big')
			|| document.querySelector('.info-data-price')
			|| document.querySelector('.price');
		let price = 0;
		if (priceEl) {
			const pm = priceEl.textContent.match(/([\d.,]+)\s*€/);
			if (pm) price = parseInt(pm[1].replace(/[.,]/g, ''), 10) || 0;
			else price = parseInt(priceEl.textContent.replace(/[^\d]/g, ''), 10) || 0;
		}

		// --- Property details from info-features section ---
		let rooms = 0, bathrooms = 0, area = 0, floor = '', hasLift = false;
		let furnished = '', deposit = '';

		// Parse the info-features list items
		const featureLis = document.querySelectorAll('.info-features .details-property_features li, .details-property li, .info-features li, .info-data .txt-big');
		featureLis.forEach(el => {
			const t = el.textContent.trim();
			const tl = t.toLowerCase();

			// Rooms: "2 bedrooms" / "2 hab." / "1 room"
			if (/\d+\s*(?:bed|hab|room|local|dormitor)/i.test(tl) && rooms === 0) {
				rooms = parseInt(tl.match(/(\d+)/)[1], 10) || 0;
			}
			// Bathrooms: "1 bathroom" / "1 baño" / "1 bagno" / "1 wc"
			if (/\d+\s*(?:bath|baño|bagn|wc|casa\s*de\s*banho)/i.test(tl) && bathrooms === 0) {
				bathrooms = parseInt(tl.match(/(\d+)/)[1], 10) || 0;
			}
			// Area: "80 m²"
			if (/[\d.,]+\s*m[²2]/i.test(tl) && area === 0) {
				const am = tl.match(/([\d.,]+)\s*m[²2]/i);
				if (am) area = parseFloat(am[1].replace(',', '.')) || 0;
			}
			// Floor: "5th floor", "Ground floor", "planta 3"
			if (/(?:ground|basement|\d+(?:st|nd|rd|th)?)\s*floor|planta\s*\w+|piano\s*\w+|andar\s*\w+/i.test(tl) && !floor) {
				const fm = tl.match(/((?:ground|basement|\d+(?:st|nd|rd|th)?)\s*floor|planta\s*\w+|piano\s*\w+|andar\s*\w+)/i);
				if (fm) floor = fm[0].trim();
			}
			// Lift
			if (/with lift|lift|ascensor|ascensore|elevador/i.test(tl)) hasLift = true;
			// Furnished
			if (/furnished|amueblado|ammobiliato|mobilado/i.test(tl)) {
				if (/unfurnished|sin amueblar|non ammobiliato|sem mobil/i.test(tl)) furnished = 'unfurnished';
				else furnished = 'furnished';
			}
			// Deposit
			if (/deposit|fianza|cauzione|cau[çc]/i.test(tl)) {
				deposit = t;
			}
		});

		// --- Also parse the info-data section for rooms/area/bathrooms ---
		document.querySelectorAll('.info-data span.txt-big, .info-data .txt-big').forEach(el => {
			const parent = el.closest('.info-data') || el.parentElement;
			const label = parent ? parent.textContent.trim().toLowerCase() : '';
			const val = parseInt(el.textContent.trim(), 10) || 0;
			if (/bed|hab|room|dormitor/i.test(label) && rooms === 0) rooms = val;
			if (/bath|baño|bagn|wc/i.test(label) && bathrooms === 0) bathrooms = val;
			if (/m[²2]/i.test(label) && area === 0) {
				area = parseFloat(el.textContent.trim().replace(',', '.')) || 0;
			}
		});

		// --- Features list (all bullet items) ---
		const features = [];
		const featureSels = [
			'.details-property_features li',
			'.details-property li',
			'.info-features li'
		];
		const seen = new Set();
		featureSels.forEach(sel => {
			document.querySelectorAll(sel).forEach(el => {
				const ft = el.textContent.trim();
				if (ft && !seen.has(ft)) {
					seen.add(ft);
					features.push(ft);
				}
			});
		});

		// --- Description ---
		const descEl = document.querySelector('.comment .adCommentsBody')
			|| document.querySelector('.adCommentsBody')
			|| document.querySelector('.expandable')
			|| document.querySelector('.comment p');
		const description = descEl ? descEl.textContent.trim() : '';

		// --- Agent ---
		const agentEl = document.querySelector('.professional-name .about-advertiser-name')
			|| document.querySelector('.about-advertiser-name')
			|| document.querySelector('.advertiser-name')
			|| document.querySelector('.professional-name');
		const agent = agentEl ? agentEl.textContent.trim() : '';

		// --- Agent phone ---
		const phoneEl = document.querySelector('.phone-btn-text')
			|| document.querySelector('.contact-phones')
			|| document.querySelector('a[href^="tel:"]');
		const agentPhone = phoneEl ? phoneEl.textContent.trim() : '';

		// --- Energy rating ---
		const energyEl = document.querySelector('.energy-certificate .energy-label')
			|| document.querySelector('.details-property_features .energy')
			|| document.querySelector('[class*="energy"] .energy-label');
		let energy = energyEl ? energyEl.textContent.trim() : '';
		// Also try the energy certificate section with the letter grade
		if (!energy) {
			const certEl = document.querySelector('.energy-certificate');
			if (certEl) {
				const activeEl = certEl.querySelector('.selected, .active, [class*="active"]');
				energy = activeEl ? activeEl.textContent.trim() : certEl.textContent.trim().substring(0, 50);
			}
		}

		// --- Images ---
		const images = [];
		const imgSeen = new Set();
		// Try multiple image container selectors
		const imgSels = [
			'.detail-multimedia img',
			'.gallery img',
			'picture img',
			'.carousel img',
			'[data-ondemand-img]',
			'.detail-image-container img'
		];
		imgSels.forEach(sel => {
			document.querySelectorAll(sel).forEach(img => {
				const src = img.src || img.dataset.src || img.dataset.ondemandImg || img.getAttribute('data-ondemand-img') || '';
				if (src && !src.includes('svg') && !src.includes('data:') && !imgSeen.has(src)) {
					imgSeen.add(src);
					images.push(src);
				}
			});
		});
		// Also check background-image on slide elements
		document.querySelectorAll('.detail-multimedia [style*="background-image"], .carousel-inner [style*="background-image"]').forEach(el => {
			const m = el.style.backgroundImage.match(/url\(["']?([^"')]+)["']?\)/);
			if (m && m[1] && !m[1].includes('svg') && !imgSeen.has(m[1])) {
				imgSeen.add(m[1]);
				images.push(m[1]);
			}
		});

		return JSON.stringify({
			id: id,
			platform: 'idealista',
			title: title,
			address: address,
			price_monthly: price,
			rooms: rooms,
			bathrooms: bathrooms,
			area_sqm: area,
			floor: floor,
			has_lift: hasLift,
			furnished: furnished,
			deposit: deposit,
			agent: agent,
			agent_phone: agentPhone,
			features: features,
			description: description,
			energy_rating: energy,
			images: images,
			url: '%s'
		});
	})()`, url, url)

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var detail PropertyDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &detail, nil
}
