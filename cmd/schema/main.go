package main

import (
	"context"
	"fmt"
	"os"

	app "github.com/youla-dev/schema/internal/app"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/joho/godotenv"
	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
)

var Version = "0.0"

func main() {
	_ = godotenv.Load(os.Getenv("SCHEMA_CONFIG"))

	app := app.NewApp(Version)
	app.Run(context.Background())
}

func getClusterClient(c *cli.Context) (*saramaCluster.Client, error) {
	cluster := c.StringSlice("cluster")
	kfkCfg := saramaCluster.NewConfig()
	clusterClient, err := saramaCluster.NewClient(cluster, kfkCfg)
	if err != nil {
		return nil, fmt.Errorf("can not create cluster client %v: %w", cluster, err)
	}
	return clusterClient, nil
}

func getSRClient(c *cli.Context) (srclient.ISchemaRegistryClient, error) {
	return srclient.CreateSchemaRegistryClient(c.String("sr")), nil
}

func subjectName(kafkaTopic, record string) string {
	return kafkaTopic + "-" + record + "-value"
}
