package data

//Config represent a data selector for projection and selection
type Config struct {
	//TODO: Should order by be a slice?
	OrderBy string `json:",omitempty"`
	Limit   int    `json:",omitempty"`
}
