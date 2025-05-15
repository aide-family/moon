package email

func copyConfig(c Config) Config {
	return &config{
		Host:   c.GetHost(),
		Port:   c.GetPort(),
		User:   c.GetUser(),
		Pass:   c.GetPass(),
		Enable: c.GetEnable(),
		Name:   c.GetName(),
	}
}

type config struct {
	Host   string `json:"host"`
	Port   uint32 `json:"port"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Enable bool   `json:"enable"`
	Name   string `json:"name"`
}

func (c *config) GetUser() string {
	return c.User
}

func (c *config) GetPass() string {
	return c.Pass
}

func (c *config) GetHost() string {
	return c.Host
}

func (c *config) GetPort() uint32 {
	return c.Port
}

func (c *config) GetEnable() bool {
	return c.Enable
}

func (c *config) GetName() string {
	return c.Name
}
