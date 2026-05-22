package ctrip

import (
	"ctrip-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type Flight struct {
	Airline       string `json:"airline"`
	FlightNo      string `json:"flight_no"`
	Aircraft      string `json:"aircraft"`
	DepTime       string `json:"dep_time"`
	DepAirport    string `json:"dep_airport"`
	ArrTime       string `json:"arr_time"`
	ArrAirport    string `json:"arr_airport"`
	Price         string `json:"price"`
	Cabin         string `json:"cabin"`
	Discount      string `json:"discount,omitempty"`
	OnTimeRate    string `json:"on_time_rate,omitempty"`
}

type FlightSearchResult struct {
	Flights []Flight `json:"flights"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	Date    string   `json:"date"`
}

func SearchFlights(client *browser.Client, from, to, date string, limit int) (*FlightSearchResult, error) {
	if date == "" {
		date = time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	}

	u := fmt.Sprintf("https://flights.ctrip.com/online/list/oneway-%s-%s?depdate=%s&cabin=y&adult=1&child=0&infant=0",
		from, to, date)

	if err := client.NavigateNewTab(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(8 * time.Second)

	js := `(function(){
		var items = document.querySelectorAll(".flight-item");
		if (!items.length) return JSON.stringify({error: "no_flights"});
		var flights = [];
		for (var i = 0; i < items.length; i++) {
			var el = items[i];
			var text = el.textContent || "";

			var airlineEl = el.querySelector(".flight-airline");
			var airline = airlineEl ? airlineEl.textContent.trim() : "";

			var detailEl = el.querySelector(".flight-detail");
			var detailText = detailEl ? detailEl.textContent.trim() : "";
			var timeMatch = detailText.match(/(\d{2}:\d{2})(.*?)(\d{2}:\d{2})/s);
			var depTime = timeMatch ? timeMatch[1] : "";
			var arrTime = timeMatch ? timeMatch[3] : "";
			var midText = timeMatch ? timeMatch[2] : detailText;
			var airports = midText.split(/\d{2}:\d{2}/);
			var depAirport = "";
			var arrAirport = "";
			var apMatch = detailText.match(/(\d{2}:\d{2})([\s\S]*?)(\d{2}:\d{2})([\s\S]*)/);
			if (apMatch) {
				depAirport = apMatch[2].trim();
				arrAirport = apMatch[4].trim();
			}

			var priceEl = el.querySelector(".flight-price");
			var price = priceEl ? priceEl.textContent.trim().replace(/\s+/g, " ") : "";

			var flightNoMatch = airline.match(/[A-Z0-9]{2}\d{3,5}/);
			var flightNo = flightNoMatch ? flightNoMatch[0] : "";

			var aircraftMatch = text.match(/(波音|空客|商飞)\S+?\(.\)/);
			var aircraft = aircraftMatch ? aircraftMatch[0] : "";

			var tagsEl = el.querySelector(".flight-tags");
			var discount = tagsEl ? tagsEl.textContent.trim() : "";

			flights.push({
				airline: airline.replace(flightNo, "").replace(aircraft, "").trim(),
				flight_no: flightNo,
				aircraft: aircraft,
				dep_time: depTime,
				dep_airport: depAirport,
				arr_time: arrTime,
				arr_airport: arrAirport,
				price: price,
				cabin: "",
				discount: discount
			});
		}
		return JSON.stringify({flights: flights});
	})()`

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var check struct {
		Error string `json:"error"`
	}
	json.Unmarshal(raw, &check)
	if check.Error != "" {
		return nil, fmt.Errorf("no flight data found (page may need more time to load)")
	}

	var parsed struct {
		Flights []Flight `json:"flights"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if limit > 0 && len(parsed.Flights) > limit {
		parsed.Flights = parsed.Flights[:limit]
	}

	return &FlightSearchResult{
		Flights: parsed.Flights,
		From:    from,
		To:      to,
		Date:    date,
	}, nil
}
