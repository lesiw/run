package main

import "net/http"

func realUrl(url string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
			if r.Response.StatusCode >= 300 && r.Response.StatusCode < 400 {
				return nil
			}
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Request.URL.String(), nil
}
