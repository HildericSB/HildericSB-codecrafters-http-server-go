package config

type Config struct {
	Port       string
	FileDir    string
	LogLevel   string
	BufferSize int
}

func DefaultConfig() Config {
	return Config{
		Port:       DEFAULT_PORT,
		FileDir:    DEFAULT_FILE_DIR,
		BufferSize: BUFFER_SIZE,
		LogLevel:   "info",
	}
}

const (
	CRLF             = "\r\n"
	BUFFER_SIZE      = 4096
	DEFAULT_FILE_DIR = "/tmp/"
	DEFAULT_PORT     = "4221"
)
