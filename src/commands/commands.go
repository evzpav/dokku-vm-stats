package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku stats[:COMMAND]

Runs commands that interact with the app's repo

Additional commands:`

	helpContent = `
    stats:start CLIENT_URL, start collecting VM data
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	mainCommand := "stats"
	subcommand := ""

	cmd := flag.Arg(0)

	fmt.Println("cmd", cmd)

	if cmd == "" || cmd == mainCommand {
		usage()
		return
	}

	if strings.Contains(cmd, ":") {
		cmds := strings.Split(cmd, ":")
		if cmds[0] != mainCommand {
			fmt.Println("invalid command")
			return
		}
		if cmds[1] == "" {
			usage()
			return
		}

		subcommand = cmds[1]
	}

	switch subcommand {
	case "help":
		usage()
	case "start":
		baseURL := flag.Arg(1)
		u, err := url.Parse(baseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			fmt.Printf("invalid url: %v\n", baseURL)
			break
		}

		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("could not get working directory: %v\n", err)
			break
		}

		everyMin := "*/1 * * * *"

		cronCmd := fmt.Sprintf("%s %s/%s --url=%s", everyMin, dir, "collectstats", baseURL)

		_, err = AddCronJob(cronCmd)
		if err != nil {
			fmt.Printf("failed to add cron command: %v\n", err)
			return
		}
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "\t"
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}
