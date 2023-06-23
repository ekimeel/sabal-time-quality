package main

type Config struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Properties  map[string]interface{} `json:"properties"`
}
