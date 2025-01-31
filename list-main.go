/*
 * MinIO Client (C) 2021 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/minio/pkg/console"
)

var allFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "auth-token, a",
		Usage: "authorization token for HCP",
		Value: "",
	},
	cli.StringFlag{
		Name:  "namespace-url, n",
		Usage: "namespace URL path, e.g https://namespace-name.tenant-name.hcp-domain-name/rest",
	},
	cli.StringFlag{
		Name:  "host-header",
		Usage: "host header for HCP",
	},
	cli.StringFlag{
		Name:  "data-dir, d",
		Usage: "path to work directory for tool",
	},
	cli.StringFlag{
		Name:  "annotation",
		Usage: "custom annotation name",
	},
	cli.BoolFlag{
		Name:  "insecure, i",
		Usage: "disable TLS certificate verification",
	},
	cli.BoolFlag{
		Name:  "log, l",
		Usage: "enable logging",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debugging",
	},
}
var (
	authToken          string
	hostHeader         string
	namespaceURL       string
	dirPath            string
	bucket             string // HCP bucket name
	minioBucket        string // in case user needs a different bucket name on MinIO
	debugFlag, logFlag bool
	annotation         string
	hcp                *hcpBackend
)

const (
	objListFile = "object_listing.txt"
	failMigFile = "migration_fails.txt"
)

var listCmd = cli.Command{
	Name:   "list",
	Usage:  "List objects in HCP namespace and download to disk",
	Action: listAction,
	Flags:  allFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} --auth-token --namespace-url --host-header --dir

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}

EXAMPLES:
1. List objects in HCP namespace https://hcp-vip.example.com and download list to /tmp/data
   $ hcp-to-minio list -a "HCP bXl1c2Vy:3f3c6784e97531774380db177774ac8d" --host-header "HOST:s3testbucket.tenant.hcp.example.com" \
  	       --namespace-url "https://hcp-vip.example.com/rest" --dir "/tmp/data"
`,
}

func checkArgsAndInit(ctx *cli.Context) {
	authToken = ctx.String("auth-token")
	hostHeader = ctx.String("host-header")
	namespaceURL = ctx.String("namespace-url")
	debugFlag = ctx.Bool("debug")
	logFlag = ctx.Bool("log")

	_, err := url.Parse(namespaceURL)
	if err != nil {
		console.Fatalln("--namespace-url malformed", namespaceURL)
	}

	dirPath = ctx.String("data-dir")
	//	bucket = ctx.String("bucket")
	annotation = ctx.String("annotation")

	if authToken == "" || hostHeader == "" || namespaceURL == "" {
		cli.ShowCommandHelp(ctx, ctx.Command.Name) // last argument is exit code
		console.Fatalln(fmt.Errorf("--auth-token, --host-header, --namespace-url and --data-dir required"))
		return
	}
	if dirPath == "" {
		console.Fatalln(fmt.Errorf("path to working dir required, please set --data-dir flag"))
		return
	}

	console.SetColor("Request", color.New(color.FgCyan))
	console.SetColor("Method", color.New(color.Bold, color.FgWhite))
	console.SetColor("Host", color.New(color.Bold, color.FgGreen))
	console.SetColor("ReqHeaderKey", color.New(color.Bold, color.FgWhite))
	console.SetColor("RespHeaderKey", color.New(color.Bold, color.FgCyan))
	console.SetColor("RespStatus", color.New(color.Bold, color.FgYellow))
	console.SetColor("ErrStatus", color.New(color.Bold, color.FgRed))
	console.SetColor("Response", color.New(color.FgGreen))
	hcp = &hcpBackend{
		URL:        namespaceURL,
		authToken:  authToken,
		hostHeader: hostHeader,
		Insecure:   ctx.Bool("insecure"),
	}
}

func listAction(cliCtx *cli.Context) error {
	checkArgsAndInit(cliCtx)
	ctx := context.Background()
	logMsg("Downloading namespace listing to disk...")
	if err := hcp.downloadObjectList(ctx); err != nil {
		logDMsg("exiting from listing", err)
		return err
	}
	return nil
}
