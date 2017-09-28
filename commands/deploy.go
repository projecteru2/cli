package commands

import (
	"io"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	enginecontainer "github.com/docker/docker/api/types/container"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
)

func RawDeploy(c *cli.Context, conn *grpc.ClientConn) {
	pod, entry, image, network, cpu, mem, envs, count := utils.GetDeployParams(c)
	if c.NArg() != 1 {
		log.Fatal("[RawDeploy] no spec")
	}
	specURI := c.Args().First()
	log.Debugf("[RawDeploy] Deploy %s", specURI)
	var data []byte
	var err error
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		log.Fatalf("[RawDeploy] read spec failed %v", err)
	}
	client := pb.NewCoreRPCClient(conn)
	opts := generateOpts(data, pod, entry, image, network, cpu, mem, envs, count)
	resp, err := client.CreateContainer(context.Background(), opts)
	if err != nil {
		log.Fatalf("[RawDeploy] send request failed %v", err)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("[RawDeploy] Message invalid %v", err)
		}

		if msg.Success {
			log.Infof("[RawDeploy] Success %s %s %s", msg.Id[:7], msg.Name, msg.Nodename)
		} else {
			log.Errorf("[RawDeploy] Failed %v", msg.Error)
		}
	}
}

func generateOpts(data []byte, pod, entry, image, network string, cpu float64, mem int64, envs []string, count int32) *pb.DeployOptions {
	coreSpecs := &coretypes.Specs{}
	if err := yaml.Unmarshal(data, &coreSpecs); err != nil {
		log.Fatalf("[generateOpts] get specs failed %v", err)
	}

	networkmode := enginecontainer.NetworkMode(network)
	networks := map[string]string{network: ""}
	if !networkmode.IsUserDefined() {
		networks = map[string]string{}
	}
	opts := &pb.DeployOptions{
		Specs:       string(data),
		Appname:     coreSpecs.Appname,
		Image:       image,
		Podname:     pod,
		Entrypoint:  entry,
		CpuQuota:    cpu,
		Memory:      mem,
		Count:       count,
		Networks:    networks,
		Networkmode: network,
		Env:         envs,
		Raw:         true,
	}
	return opts
}
