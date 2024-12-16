package main

type API struct {
	URL    string      `json:"url" binding:"required"`
	Status int         `json:"status" binding:"required"`
	Method string      `json:"method" binding:"required"`
	Data   interface{} `json:"data" binding:"required"`
}

type Result struct {
	Data   string `json:"data"`
	Status int    `json:"status"`
}

type Response struct {
	Results []Result `json:"results"`
}
