package boss

import (
	"boss-cli/browser"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Job struct {
	JobName    string   `json:"job_name"`
	Salary     string   `json:"salary"`
	Experience string   `json:"experience"`
	Degree     string   `json:"degree"`
	City       string   `json:"city"`
	District   string   `json:"district"`
	Area       string   `json:"area"`
	Company    string   `json:"company"`
	Industry   string   `json:"industry"`
	Scale      string   `json:"scale"`
	Stage      string   `json:"stage"`
	Skills     []string `json:"skills"`
	Welfare    []string `json:"welfare"`
	BossName   string   `json:"boss_name"`
	BossTitle  string   `json:"boss_title"`
	BossOnline bool     `json:"boss_online"`
	JobID      string   `json:"job_id"`
	DetailURL  string   `json:"detail_url"`
}

type SearchResult struct {
	Jobs  []Job `json:"jobs"`
	Total int   `json:"total"`
}

func SearchJobs(client *browser.Client, query, city, salary, experience, degree, scale, stage, jobType string, page, limit int) (*SearchResult, error) {
	if city == "" {
		city = "101010100"
	}

	u := fmt.Sprintf("https://www.zhipin.com/web/geek/job?query=%s&city=%s",
		url.QueryEscape(query), url.QueryEscape(city))

	if salary != "" {
		u += "&salary=" + url.QueryEscape(salary)
	}
	if experience != "" {
		u += "&experience=" + url.QueryEscape(experience)
	}
	if degree != "" {
		u += "&degree=" + url.QueryEscape(degree)
	}
	if scale != "" {
		u += "&scale=" + url.QueryEscape(scale)
	}
	if stage != "" {
		u += "&stage=" + url.QueryEscape(stage)
	}
	if jobType != "" {
		u += "&jobType=" + url.QueryEscape(jobType)
	}

	if err := client.Navigate(u); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(4 * time.Second)

	// Load additional pages by calling Vue's internal loadMore
	if page > 1 {
		for i := 1; i < page; i++ {
			loadJS := `(function(){
				var el = document.querySelector(".page-jobs-main");
				if (!el || !el.__vue__) return "no_vue";
				var v = el.__vue__;
				v.$data.moreLoading = false;
				v.$data.hasMore = true;
				v.loadMoreHandle();
				return "ok";
			})()`
			client.EvaluateJSON(loadJS)
			time.Sleep(2 * time.Second)
		}
	}

	js := `(function(){
		var cards = document.querySelectorAll(".job-card-wrap");
		var jobs = [];
		for (var i = 0; i < cards.length; i++) {
			var v = cards[i].__vue__;
			if (!v || !v.$props || !v.$props.data) continue;
			var d = v.$props.data;
			jobs.push({
				job_name: d.jobName || "",
				salary: d.salaryDesc || "",
				experience: d.jobExperience || "",
				degree: d.jobDegree || "",
				city: d.cityName || "",
				district: d.areaDistrict || "",
				area: d.businessDistrict || "",
				company: d.brandName || "",
				industry: d.brandIndustry || "",
				scale: d.brandScaleName || "",
				stage: d.brandStageName || "",
				skills: d.skills || [],
				welfare: d.welfareList || [],
				boss_name: d.bossName || "",
				boss_title: d.bossTitle || "",
				boss_online: !!d.bossOnline,
				job_id: d.encryptJobId || "",
				detail_url: "https://www.zhipin.com/job_detail/" + (d.encryptJobId || "") + ".html"
			});
		}
		return JSON.stringify({jobs: jobs, total: jobs.length});
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
		return nil, fmt.Errorf("page data not available: %s", check.Error)
	}

	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if limit > 0 && len(result.Jobs) > limit {
		result.Jobs = result.Jobs[:limit]
		result.Total = limit
	}

	return &result, nil
}
