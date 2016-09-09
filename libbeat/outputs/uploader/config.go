package uploader

import "time"

type uploaderConfig struct {
	Hosts  []string `config:"hosts"`
	Network	string	`config:"network"`
	Cid	string	`config:"cid"`
	HostIp	string	`config:"ip"`
	ConnectTimeout   time.Duration	`config:"connectTimeout"`
	WriteTimeout    time.Duration	`config:"writeTimeout"`
}


var (
	defaultConfig = uploaderConfig{
		Network: "tcp",
		ConnectTimeout: 5*time.Second,
		WriteTimeout: 5*time.Second,
	}
)




