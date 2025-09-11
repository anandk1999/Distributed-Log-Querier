package common

type ServerRequest struct {
	Input    string `json:"input"`
	FileType string `json:"file_type"`
}

type ServerResponse struct {
	Output  string `json:"output"`
	LogFile string `json:"log_file"`
}
