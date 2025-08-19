package app

type TrivyRepo struct {
	Trivy    string `json:"trivy"     required:"true"`
	TrivyDB  string `json:"trivy_db"  required:"true"`
	VulnList string `json:"vuln_list" required:"true"`
}

func (t *TrivyRepo) SetDefault() {
	if t.Trivy == "" {
		t.Trivy = "https://github.com/wjunLu/trivy.git"
	}

	if t.TrivyDB == "" {
		t.TrivyDB = "https://github.com/wjunLu/trivy-db.git"
	}

	if t.VulnList == "" {
		t.VulnList = "https://github.com/aquasecurity/vuln-list.git"
	}
}

type Concurrency struct {
	Num int `json:"num" required:"true"`
}

func (c *Concurrency) SetDefault() {
	if c.Num == 0 {
		c.Num = 10
	}
}
