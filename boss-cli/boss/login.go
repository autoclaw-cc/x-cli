package boss

import (
	"boss-cli/browser"
	"encoding/json"
	"fmt"
	"time"
)

type LoginStatus struct {
	LoggedIn bool   `json:"logged_in"`
	UserName string `json:"username"`
}

func CheckLogin(client *browser.Client) (*LoginStatus, error) {
	if err := client.Navigate("https://www.zhipin.com/web/geek/job?query=test&city=101010100"); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	time.Sleep(3 * time.Second)

	js := `(function(){
		var navItems = document.querySelectorAll(".user-nav a, .nav-userinfo a");
		var userName = "";
		var isLogin = false;
		var avatar = document.querySelector(".user-nav .nav-figure img");
		if (avatar) {
			isLogin = true;
		}
		var loginBtn = document.querySelector(".btn-login, .header-login-btn");
		if (loginBtn) {
			isLogin = false;
		}
		for (var i = 0; i < navItems.length; i++) {
			var text = navItems[i].textContent.trim();
			var href = navItems[i].getAttribute("href") || "";
			if (href.indexOf("/user/") > -1 || (text.length > 0 && text.length < 10 && text.indexOf("登录") === -1 && text.indexOf("注册") === -1 && text.indexOf("消息") === -1 && text.indexOf("简历") === -1 && text.indexOf("APP") === -1 && text.indexOf("VIP") === -1 && text.indexOf("规则") === -1 && text.indexOf("账号") === -1 && text.indexOf("隐私") === -1 && text.indexOf("通知") === -1 && text.indexOf("切换") === -1 && text.indexOf("退出") === -1)) {
				userName = text;
				break;
			}
		}
		return JSON.stringify({logged_in: isLogin, username: userName});
	})()`

	raw, err := client.EvaluateJSON(js)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	var status LoginStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return &status, nil
}
