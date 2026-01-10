package credential

import "fmt"

type AppConfig struct {
    Application ApplicationConfig `mapstructure:"application"`
    Database    DatabaseConfig    `mapstructure:"db"`
}

// ApplicationConfig holds application-level config
type ApplicationConfig struct {
    Name string `mapstructure:"name"`
    Mode string `mapstructure:"mode"`
    Env  string `mapstructure:"env"`
    Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
    Configs DBConnectionConfig `mapstructure:"configs"`
}

type DBConnectionConfig struct {
    Host        string `mapstructure:"host"`
    Port        string `mapstructure:"port"`
    User        string `mapstructure:"user"`
    Password    string `mapstructure:"password"`
    Name        string `mapstructure:"name"`
    SSLMode     string `mapstructure:"sslmode"`
    MaxIdleConn int    `mapstructure:"maxIdleConn"`
    MaxOpenConn int    `mapstructure:"maxOpenConn"`
}


// Helper methods for convenience
func (c *AppConfig) GetDSNPostgreSQL() string {
    db := c.Database.Configs
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        db.User,
        db.Password,
        db.Host,
        db.Port,
        db.Name,
        db.SSLMode,
    )
}
