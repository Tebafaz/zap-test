package logger

type (
	Config struct {
		Cores       []Core     `yaml:"cores"`
		Caller      bool       `yaml:"caller"`
		Development bool       `yaml:"development"`
		Udp         string     `yaml:"udp"`
		FileConfig  FileConfig `yaml:"fileConfig"`
	}
	Core struct {
		Addr         string       `yaml:"addr"`
		Host         string       `yaml:"host"`
		Level        string       `yaml:"level"`
		Encoding     string       `yaml:"encoding"`
		Stacktrace   string       `yaml:"stacktrace"`
		OutputConfig OutputConfig `yaml:"outputConfig"`
	}
	OutputConfig struct {
		UseCapitalColor bool   `yaml:"useCapitalColor"`
		UseStdOut       bool   `yaml:"useStdOut"`
		Path            string `yaml:"path"`
		FlushSeconds    int64  `yaml:"flushseconds"`
		BufferSize      int    `yaml:"buffersize"`
	}

	// FileConfig deprecated: use OutputConfig instead
	FileConfig struct {
		Path         string
		FlushSeconds int64
		BufferSize   int
	}
)

func NewDefaultConfig(logFile string) *Config {
	return &Config{
		Cores: []Core{
			{
				Addr:       "",
				Host:       "",
				Level:      "dpanic",
				Encoding:   "console",
				Stacktrace: "error",
				OutputConfig: OutputConfig{
					UseCapitalColor: true,
					UseStdOut:       true,
					Path:            "",
					FlushSeconds:    0,
					BufferSize:      0,
				},
			},
			{
				Addr:       "",
				Host:       "",
				Level:      "debug",
				Encoding:   "json",
				Stacktrace: "error",
				OutputConfig: OutputConfig{
					UseCapitalColor: false,
					UseStdOut:       false,
					Path:            logFile,
					FlushSeconds:    5,
					BufferSize:      512 * 1024,
				},
			},
		},
		Caller:      true,
		Development: false,
		Udp:         "",
	}
}
