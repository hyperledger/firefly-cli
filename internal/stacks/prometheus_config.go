package stacks

import "fmt"

type GlobalConfig struct {
	ScrapeInterval string `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout string `yaml:"scrape_timeout,omitempty"`
}

type ScrapeConfig struct {
	JobName string `yaml:"job_name,omitempty"`
	MetricsPath string `yaml:"metrics_path,omitempty"`
	StaticConfigs []*StaticConfig `yaml:"static_configs,omitempty"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets,omitempty"`
}

type PrometheusConfig struct {
	Global *GlobalConfig  `yaml:"global,omitempty"`
	ScrapeConfigs []*ScrapeConfig `yaml:"scrape_configs,omitempty"`
}

func (s *StackManager) GeneratePrometheusConfig() *PrometheusConfig {
	config := &PrometheusConfig{
		Global: &GlobalConfig{
			ScrapeInterval: "5s",
			ScrapeTimeout: "5s",
		},
		ScrapeConfigs: []*ScrapeConfig{
			{
				JobName: "fireflies",
				MetricsPath: "/metrics",
				StaticConfigs: []*StaticConfig{
					{
						Targets: []string{},
					},
				},
			},
		},
	}

	for i, member := range s.Stack.Members {
		config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets, fmt.Sprintf("firefly_core_%d:%d", i, member.ExposedFireflyMetricsPort))
	}

	return config
}