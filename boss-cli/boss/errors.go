package boss

import "errors"

var ErrNotLoggedIn = errors.New("not logged in to Boss直聘. Please open Chrome, navigate to https://www.zhipin.com, and log in manually")
