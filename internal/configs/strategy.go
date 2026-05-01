package configs

import "gopkg.in/yaml.v3"

type StrategyFeature struct {
	AlertQuota AlertQuota `yaml:"alert_quota"`
}

// AlertQuota holds per-plan alert creation limits.
// Max == 0 means unlimited.
type AlertQuota struct {
	Free int `yaml:"free"`
	Pro  int `yaml:"pro"`
	Max  int `yaml:"max"`
}

// UnmarshalYAML decodes alert_quota from YAML.
// For the Max field, any non-integer value (e.g. "*") is treated as 0 (unlimited).
func (q *AlertQuota) UnmarshalYAML(node *yaml.Node) error {
	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]
		switch key {
		case "free":
			if err := val.Decode(&q.Free); err != nil {
				return err
			}
		case "pro":
			if err := val.Decode(&q.Pro); err != nil {
				return err
			}
		case "max":
			if err := val.Decode(&q.Max); err != nil {
				q.Max = 0 // non-integer (e.g. "*") → unlimited
			}
		}
	}
	return nil
}

func GetConfigStrategy(c Config) StrategyFeature {
	return c.Strategy
}
