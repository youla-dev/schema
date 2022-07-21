package main

import (
	"fmt"
	"log"
	"os"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/joho/godotenv"
	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
)

const (
	cmdRegister = "register"
	cmdDelete   = "delete"
	cmdValidate = "validate"
	cmdVersions = "versions"
	cmdInspect  = "inspect"
	cmdSubjects = "subjects"
	cmdExport   = "export"
)

var Version = "0.0"

var (
	flagClusterRequired = &cli.StringSliceFlag{
		Name:     "cluster",
		Required: true,
		Value:    cli.NewStringSlice("localhost:9092"),
		Usage:    "List of Kafka brokers joined with comma.",
		EnvVars:  []string{"CLUSTER"},
	}
	flagSRRequired = &cli.StringFlag{
		Name:     "sr",
		Required: true,
		Value:    "http://localhost:8081",
		Usage:    "URL for the Confluent Schema Registry.",
		EnvVars:  []string{"SCHEMA_REGISTRY"},
	}
	flagSR = &cli.StringFlag{
		Name:    "sr",
		Value:   "http://localhost:8081",
		Usage:   "URL for the Confluent Schema Registry.",
		EnvVars: []string{"SCHEMA_REGISTRY"},
	}
	flagTopicRequired = &cli.StringFlag{
		Name:     "topic",
		Required: true,
		Value:    "test-topic",
		Usage:    "Name of the topic.",
		EnvVars:  []string{"TOPIC"},
	}
	flagTopic = &cli.StringFlag{
		Name:    "topic",
		Usage:   "Name of the topic.",
		EnvVars: []string{"TOPIC"},
	}
	flagRecord = &cli.StringFlag{
		Name:    "record",
		Value:   "default",
		Usage:   "Name of the record within the topic.",
		EnvVars: []string{"RECORD"},
	}
	flagPermanent = &cli.BoolFlag{
		Name:    "permanent",
		Value:   false,
		Usage:   "Permanent flag is used for the removing of version in 'hard' mode.",
		EnvVars: []string{"PERMANENT"},
	}
	flagProtoRequired = &cli.StringFlag{
		Name:     "proto",
		Required: true,
		Usage:    "File with Protobuf schema definition.",
		EnvVars:  []string{"PROTO"},
	}
	flagVersion = &cli.StringFlag{
		Name:    "version",
		Value:   "latest",
		Usage:   "Version of the schema. E.g. `latest`, `1`, etc.",
		EnvVars: []string{"VERSION"},
	}
)

func main() {
	_ = godotenv.Load(os.Getenv("SCHEMA_CONFIG"))

	app := &cli.App{
		Name:    "Schema Registry utility",
		Version: Version,
		Usage:   "Utility for your CI/CD process to validate, register or delete Kafka protobuf schemes in the registry.",
		Commands: []*cli.Command{
			{
				Name:      cmdRegister,
				Usage:     "Creates a subject, if one does not exists, sets a scheme for a subject or updates it.",
				ArgsUsage: "Set both the kafka cluster and the registry addresses and the protobuf file with the schema and topic&record options or set topic and record values with the separate flags.",
				Action:    register,
				Flags: []cli.Flag{
					flagClusterRequired,
					flagSRRequired,
					flagTopic,
					flagRecord,
					flagProtoRequired,
				},
			},
			{
				Name:      cmdDelete,
				Usage:     "Deletes the specified version of the schema or deletes the latest version of the scheme by default.",
				ArgsUsage: "Set the topic, record and version of the schema to delete.",
				Action:    deleteSubject,
				Flags: []cli.Flag{
					flagSRRequired,
					flagTopicRequired,
					flagRecord,
					flagVersion,
					flagPermanent,
				},
			},
			{
				Name:   cmdValidate,
				Usage:  "Validates the topic to exist and the schema changes compatibility with existing version. The schema is also valid if the the topic or subject does not exists.",
				Action: validate,
				Flags: []cli.Flag{
					flagClusterRequired,
					flagSRRequired,
					flagTopic,
					flagRecord,
					flagProtoRequired,
				},
			},
			{
				Name:   cmdVersions,
				Usage:  "Lists available versions for the subject.",
				Action: versions,
				Flags: []cli.Flag{
					flagSRRequired,
					flagTopicRequired,
					flagRecord,
				},
			},
			{
				Name:   cmdInspect,
				Usage:  "Outputs all information about the subject.",
				Action: inspect,
				Flags: []cli.Flag{
					flagSRRequired,
					flagTopicRequired,
					flagRecord,
					flagVersion,
				},
			},
			{
				Name:   cmdSubjects,
				Usage:  "Lists available records for the topic.",
				Action: subjects,
				Flags: []cli.Flag{
					flagSRRequired,
					flagTopicRequired,
				},
			},
			{
				Name:   cmdExport,
				Usage:  "Exports the schema value to the local file.",
				Action: export,
				Flags: []cli.Flag{
					flagSRRequired,
					flagTopicRequired,
					flagRecord,
					flagVersion,
					&cli.StringFlag{
						Name:     "output",
						Required: true,
						Usage:    "Output file for the scheme from registry.",
						EnvVars:  []string{"OUTPUT"},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
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
