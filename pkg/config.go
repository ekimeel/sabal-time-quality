package main

type Config struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Data        map[string]interface{} `json:"data"`
}
