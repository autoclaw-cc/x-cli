package boss

import (
	"boss-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type JobDetail struct {
	JobName     string   `json:"job_name"`
	Salary      string   `json:"salary"`
	City        string   `json:"city"`
	Experience  string   `json:"experience"`
	Degree      string   `json:"degree"`
	Skills      []string `json:"skills"`
	Description string   `json:"description"`
	Company     string   `json:"company"`
	CompanyInfo string   `json:"company_info"`
	Address     string   `json:"address"`
	BossName    string   `json:"boss_name"`
	BossTitle   string   `json:"boss_title"`
	JobID       string   `json:"job_id"`
	URL         string   `json:"url"`
}

func GetJobDetail(client *browser.Client, jobID string) (*JobDetail, error) {
	detailURL := fmt.Sprintf("https://www.zhipin.com/job_detail/%s.html", jobID)

	if err := client.NavigateNewTab(detailURL); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(3 * time.Second)

	js := `(function(){
		var result = {};

		// Get basic info from global _jobInfo
		if (typeof _jobInfo !== "undefined") {
			result.job_id = _jobInfo.job_id || "";
			result.job_name = _jobInfo.job_name || "";
			result.salary = _jobInfo.job_salary || "";
			result.company = _jobInfo.company || "";
		}

		// Get job name and salary from DOM as fallback
		var nameEl = document.querySelector(".name h1");
		if (nameEl && !result.job_name) {
			result.job_name = nameEl.innerText.trim().split("\n")[0].trim();
		}
		var salaryEl = document.querySelector(".salary");
		if (salaryEl && !result.salary) {
			result.salary = salaryEl.innerText.trim();
		}

		// Get description (innerText respects CSS visibility, filters anti-scraping)
		var descEl = document.querySelector(".job-sec-text");
		result.description = descEl ? descEl.innerText.trim() : "";

		// Get skills - use innerText to filter anti-scraping hidden text
		var skillEls = document.querySelectorAll(".job-detail-section .job-tags li, .job-keyword-list li");
		var skills = [];
		for (var i = 0; i < skillEls.length; i++) {
			var t = skillEls[i].innerText.trim();
			if (t && t.length < 30) skills.push(t);
		}
		result.skills = skills;

		// Get address
		var locEl = document.querySelector(".location-address");
		result.address = locEl ? locEl.innerText.trim() : "";

		// Parse experience and degree from the info banner
		var infoEl = document.querySelector(".job-banner .info-primary p, .text-desc");
		if (infoEl) {
			var infoText = infoEl.innerText.trim();
			var parts = infoText.split(/\s+/);
			for (var i = 0; i < parts.length; i++) {
				var p = parts[i].trim();
				if (!p) continue;
				if (/\d+-\d+年|年以[上内]|经验不限|在校|应届/.test(p)) {
					result.experience = p;
				} else if (/专|本科|硕士|博士|高中|中技|初中|学历/.test(p)) {
					result.degree = p;
				} else if (/^[一-鿿]{2,4}$/.test(p) && !result.city) {
					result.city = p;
				}
			}
		}

		// Fallback city from address
		if (!result.city && result.address) {
			var m = result.address.match(/^([一-鿿]{2,3})/);
			if (m) result.city = m[1];
		}

		// Get company info section
		var companySections = document.querySelectorAll(".job-detail-section");
		for (var i = 0; i < companySections.length; i++) {
			var sec = companySections[i];
			if (sec.className.indexOf("job-detail-company") > -1) {
				var companyText = sec.querySelector(".fold-text, .text");
				result.company_info = companyText ? companyText.innerText.trim() : sec.innerText.substring(0, 1000).trim();
				break;
			}
		}

		// Get boss info - name is in h2.name under .job-boss-info, title in .boss-info-attr
		var bossNameEl = document.querySelector(".job-boss-info h2.name");
		if (bossNameEl) {
			result.boss_name = bossNameEl.childNodes[0].textContent.trim();
		}
		var bossAttr = document.querySelector(".job-boss-info .boss-info-attr");
		if (bossAttr) {
			result.boss_title = bossAttr.innerText.trim().replace(/\s+/g, " ");
		}

		result.url = window.location.href;

		return JSON.stringify(result);
	})()`

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var detail JobDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if detail.JobID == "" {
		detail.JobID = jobID
	}

	return &detail, nil
}
