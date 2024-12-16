package main

type API struct {
	URL  string      `json:"url" binding:"required"`
	Data interface{} `json:"data" binding:"required"`
}
