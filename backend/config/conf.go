package config

var Config *Conf

type Conf struct {
	Mysql      MysqlConfig      `toml:"mysql"`
	Redis      RedisConfig      `toml:"redis"`
	Blockchain BlockchainConfig `toml:"blockchain"`

	Email        EmailConfig        `toml:"email"`
	DefaultAdmin DefaultAdminConfig `toml:"default_admin"`
	Jwt          JwtConfig          `toml:"jwt"`
	Env          EnvConfig          `toml:"env"`
}

type EmailConfig struct {
	Username string   `toml:"username"`
	Pwd      string   `toml:"pwd"`
	Host     string   `toml:"host"`
	Port     string   `toml:"port"`
	From     string   `toml:"from"`
	Subject  string   `toml:"subject"`
	To       []string `toml:"to"`
	Cc       []string `toml:"cc"`
}

type DefaultAdminConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type JwtConfig struct {
	SecretKey  string `toml:"secret_key"`
	ExpireTime int    `toml:"expire_time"` // duration, s
}

type EnvConfig struct {
	Port               string `toml:"port"`
	Version            string `toml:"version"`
	Protocol           string `toml:"protocol"`
	DomainName         string `toml:"domain_name"`
	TaskDuration       int64  `toml:"task_duration"`
	WssTimeoutDuration int64  `toml:"wss_timeout_duration"`
	TaskExtendDuration int64  `toml:"task_extend_duration"`
}

type MysqlConfig struct {
	Host         string `toml:"host"`
	Port         string `toml:"port"`
	DbName       string `toml:"dbname"`
	UserName     string `toml:"username"`
	Password     string `toml:"password"`
	MaxOpenConns int    `toml:"max_open_conns"`
	MaxIdleConns int    `toml:"max_idle_conns"`
	MaxLifeTime  int    `toml:"max_life_time"`
}

type RedisConfig struct {
	Host        string `toml:"host"`
	Port        string `toml:"port"`
	DB          int    `toml:"db"`
	Password    string `toml:"password"`
	MaxIdle     int    `toml:"max_idle"`
	MaxActive   int    `toml:"max_active"`
	IdleTimeout int    `toml:"idle_timeout"`
}

type BlockchainConfig struct {
	LocalRPCURL       string `toml:"local_rpc_url"`
	MainnetRPCURL     string `toml:"mainnet_rpc_url"`
	ScanIntervalSecs  int    `toml:"scan_interval_seconds"`
	MiniVaultAddress  string `toml:"mini_vault_address"`
	MiniMUSDAddress   string `toml:"mini_musd_address"`
	MiniStETHAddress  string `toml:"mini_steth_address"`
	MiniOracleAddress string `toml:"mini_oracle_address"`
}
