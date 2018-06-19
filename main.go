package main

import (
	"fmt"
	"os"
	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "s3 plugin"
	app.Usage = "s3 plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for the s3 connection",
			EnvVar: "PLUGIN_ENDPOINT,S3_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "aws bucket",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_BUCKET,S3_BUCKET",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_REGION,S3_REGION",
		},
		cli.StringFlag{
			Name:   "acl",
			Usage:  "upload files with acl",
			Value:  "private",
			EnvVar: "PLUGIN_ACL",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "upload files from source folder",
			EnvVar: "PLUGIN_SOURCE",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "upload files to target folder",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringFlag{
			Name:   "strip-prefix",
			Usage:  "strip the prefix from the target",
			EnvVar: "PLUGIN_STRIP_PREFIX",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "ignore files matching exclude pattern",
			EnvVar: "PLUGIN_EXCLUDE",
		},
		cli.StringFlag{
			Name:   "encryption",
			Usage:  "server-side encryption algorithm, defaults to none",
			EnvVar: "PLUGIN_ENCRYPTION",
		},
		cli.BoolTFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "PLUGIN_DRY_RUN",
		},
		cli.BoolTFlag{
			Name:   "path-style",
			Usage:  "use path style for bucket paths",
			EnvVar: "PLUGIN_PATH_STYLE",
		},
		cli.BoolTFlag{
			Name:   "yaml-verified",
			Usage:  "Ensure the yaml was signed",
			EnvVar: "DRONE_YAML_VERIFIED",
		},
		cli.BoolTFlag{
			Name:   "create-bucket-if-necessary",
			Usage:  "Create bucket if non existing yet",
			EnvVar: "PLUGIN_CREATE_BUCKET",
		},
		cli.BoolTFlag{
			Name:   "append-branch-to-bucket",
			Usage:  "Append BranchName to BucketName",
			EnvVar: "PLUGIN_APPEND_BRANCH",
		},
		cli.StringFlag{
			Name:   "prefix-delete",
			Usage:  "rm prefix in BranchName for BucketName if append-branch-name-to-bucket is true",
			EnvVar: "PLUGIN_PREFIX_RM, BRANCH_PREFIX_RM",
		},
		cli.StringFlag{
			Name:   "commit-branch",
			Usage:  "Commit Branch Name",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.BoolTFlag{
			Name:   "s3-hosting",
			Usage:  "Check and Activate S3 Hosting if not",
			EnvVar: "PLUGIN_HOSTING,S3_HOSTING",
		},
		cli.StringFlag{
			Name:   "index-document",
			Usage:  "Check and set IndexDocument in S3 Hosting Config",
			EnvVar: "PLUGIN_INDEX_DOCUMENT,S3_INDEX_DOCUMENT",
		},
		cli.StringFlag{
			Name:   "error-document",
			Usage:  "Check and set ErrorDocument in S3 Hosting Config",
			EnvVar: "PLUGIN_ERROR_DOCUMENT,S3_ERROR_DOCUMENT",
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		Endpoint:     c.String("endpoint"),
		Key:          c.String("access-key"),
		Secret:       c.String("secret-key"),
		Bucket:       c.String("bucket"),
		Region:       c.String("region"),
		Access:       c.String("acl"),
		Source:       c.String("source"),
		Target:       c.String("target"),
		StripPrefix:  c.String("strip-prefix"),
		Exclude:      c.StringSlice("exclude"),
		Encryption:   c.String("encryption"),
		PathStyle:    c.Bool("path-style"),
		DryRun:       c.Bool("dry-run"),
		YamlVerified: c.BoolT("yaml-verified"),
		CreateBucketIfNecessary: c.Bool("create-bucket-if-necessary"),
		AppendBranchtoBucket: c.Bool("append-branch-to-bucket"),
		PrefixDelete: c.String("prefix-delete"),
		CommitBranch: c.String("commit-branch"),
		S3Hosting: c.Bool("s3-hosting"),
		IndexDocument: c.String("index-document"),
		ErrorDocument: c.String("error-document"),
	}

	return plugin.Exec()
}
