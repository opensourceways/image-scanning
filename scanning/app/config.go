package app

type TrivyRepo struct {
	Trivy    string `json:"trivy"     required:"true"`
	TrivyDB  string `json:"trivy_db"  required:"true"`
	VulnList string `json:"vuln_list" required:"true"`
}
