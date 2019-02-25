package config

type Returners struct {
	Mysql
	Mongo
}

type Mysql struct {
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
	DB     string `json:"db"`
}

type Mongo struct {
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Passwd     string `json:"passwd"`
	DB         string `json:"db"`
	AuthSource string `json:"authentication_source"`
}

type MasterOptions struct {
	BaseDir           string `json:"base_dir"`
	Mode              string `json:"mode"`
	ID                string `json:"id"`
	PublicIp          string `json:"public_ip"`
	Region            string `json:"region"`
	ListenHost        string `json:"listen_host"`
	RetPort           int    `json:"ret_port"`
	Transport         string `json:"transport"`
	SockDir           string `json:"sock_dir"`
	PublishPort       int    `json:"PublishPort"`
	ProxyPort         int    `json:"proxy_port"`
	MaxOpenFile       int    `json:"max_open_file"`
	WorkerThread      int    `json:"worker_thread"`
	PkiDir            string `json:"pki_dir"`
	User              string `json:"user"`
	KeySize           int    `json:"key_size"`
	EvenReturn        string `json:"even_return"`
	TCPKeepAlive      bool   `json:"tcp_keepalive"`
	TcpKeepAliveCnt   int    `json:"tcp_keepalive_cnt"`
	TcpKeepAliveIntvl int    `json:"tcp_keepalive_intvl"`
	TcpKeepAliveIdle  int    `json:"tcp_keepalive_idle"`
	PingInterval      int    `json:"ping_interval"`
	JobBroker         string `json:"job_broker"`
	JobsCache         string `json:"jobs_cache"`
	CacheDir          string `json:"cache_dir"`
	BtSavePath        string `json:"bt_save_path"`
	BtClientPort      int    `json:"bt_client_port"`
	BtSSLPort         int    `json:"bt_ssl_port"`
	BtAnnouce         []string
	TimeOut           int
	Returner          Returners //
	ZKCluster         []string  `json:"zk_cluster"`
	LogDir            string
	Debug             bool
}

type Master struct {
	MasterID string `json:"master_ip"`
	MasterIP string `json:"master_id"`
}

type MinionOptions struct {
	BaseDir           string   `json:"base_dir"`
	Mode              string   `json:"mode"`
	ID                string   `json:"id"`
	Masters           []Master `json:"masters"`
	MasterIP          string   `json:"master_ip"`
	Region            string   `json:"region"`
	ListenHost        string   `json:"listen_host"`
	RetPort           int      `json:"ret_port"`
	Transport         string   `json:"transport"`
	SockDir           string   `json:"sock_dir"`
	ProcDir           string   `json:"proc_dir"`
	PublishPort       int      `json:"PublishPort"`
	ProxyPort         int      `json:"proxy_port"`
	MaxOpenFile       int      `json:"max_open_file"`
	WorkerThread      int      `json:"worker_thread"`
	PkiDir            string   `json:"pki_dir"`
	User              string   `json:"user"`
	KeySize           int      `json:"key_size"`
	EvenReturn        string   `json:"even_return"`
	TCPKeepAlive      bool     `json:"tcp_keepalive"`
	TcpKeepAliveCnt   int      `json:"tcp_keepalive_cnt"`
	TcpKeepAliveIntvl int      `json:"tcp_keepalive_intvl"`
	TcpKeepAliveIdle  int      `json:"tcp_keepalive_idle"`
	PingInterval      int      `json:"ping_interval"`
	JobsCache         string   `json:"jobs_cache"`
	CacheDir          string   `json:"cache_dir"`
	BtSavePath        string   `json:"bt_save_path"`
	BtClientPort      int      `json:"bt_client_port"`
	BtSSLPort         int      `json:"bt_ssl_port"`
	BtAnnouce         []string
	TimeOut           int `json:"TimeOut"`
	LogDir            string
	Debug             bool
}
