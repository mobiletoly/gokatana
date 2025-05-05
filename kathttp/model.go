package kathttp

type Version struct {
	Healthy bool   `json:"healthy"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type Page struct {
	Offset int64
	Limit  int64
}
