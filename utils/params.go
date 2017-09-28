package utils

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
)

func GetDeployParams(c *cli.Context) (string, string, string, string, float64, int64, []string, int32) {
	pod := c.String("pod")
	entry := c.String("entry")
	image := c.String("image")
	network := c.String("network")
	cpu := c.Float64("cpu")
	mem := c.Int64("mem")
	envs := c.StringSlice("env")
	count := int32(c.Int("count"))
	if pod == "" || entry == "" || image == "" {
		log.Fatal("[RawDeploy] no pod or entry or image")
	}
	return pod, entry, image, network, cpu, mem, envs, count
}
