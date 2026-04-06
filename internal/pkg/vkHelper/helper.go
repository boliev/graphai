package vkHelper

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/vkapps"
)

func GetVKUserID(r *http.Request, clientSecret string) (int64, error) {
	pv := vkapps.NewParamsVerification(clientSecret)
	u := &url.URL{RawQuery: r.URL.RawQuery}
	ok, err := pv.Verify(u)
	if err != nil || !ok {
		return 0, fmt.Errorf("invalid vk signature %v", err)
	}

	vkUserIDRaw := r.URL.Query().Get("vk_user_id")
	if vkUserIDRaw == "" {
		return 0, fmt.Errorf("vk_user_id is require %v", err)
	}

	vkUserID, err := strconv.ParseInt(vkUserIDRaw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid vk_user_id %v", err)
	}

	return vkUserID, nil
}
