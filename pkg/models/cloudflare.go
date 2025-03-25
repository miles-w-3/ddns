package models

type CloudflareResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CloudflareZonesResponse struct {
	Success  bool                     `json:"success"`
	Errors   []CloudflareResponseInfo `json:"errors"`
	Messages []CloudflareResponseInfo `json:"messages"`
	Result   []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

type CloudflareRecordResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CloudflareRecordsResponse struct {
	Success  bool                       `json:"success"`
	Errors   []CloudflareResponseInfo   `json:"errors"`
	Messages []CloudflareResponseInfo   `json:"messages"`
	Result   []CloudflareRecordResponse `json:"result"`
}
