package config

var Conf = struct {
	Port            string `yaml:"port"`
	ProxyPort       int    `yaml:"proxy_port"`
	HTTPPort        string `yaml:"http_port"`
	UploadPath      string `yaml:"upload_path"`
	UploadUrl       string `yaml:"upload_url"`
	UploadAssetUrl  string `yaml:"upload_asset_url"`
	UploadVideoUrl  string `yaml:"upload_video_url"`
	UploadOfficeUrl string `yaml:"upload_office_url"`
	AppId           string `yaml:"app_id"`
	AppSecret       string `yaml:"app_secret"`
	CanUseVrMuseum  int    `yaml:"can_use_vr_museum"`
	Db              struct {
		Mongo struct {
			Hosts    []string `yaml:"hosts"`
			User     string   `yaml:"user"`
			Pwd      string   `yaml:"pwd"`
			Database string   `yaml:"database"`
		} `yaml:"mongo"`
		Redis struct {
			Host     string `yaml:"host"`
			Pwd      string `yaml:"pwd"`
			Database int    `yaml:"database"`
		} `yaml:"redis"`
	} `yaml:"db"`
}{}

// func init() {
// 	var confFile = "config.yaml"
// 	yamlFile, err := os.ReadFile(confFile)
// 	if err != nil {
// 		log.Println("配置文件", confFile, " 不存在")
// 		return
// 	}

// 	yaml.Unmarshal(yamlFile, &Conf)
// }
