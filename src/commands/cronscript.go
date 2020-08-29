package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func AddCronJob(cronCmd string) (string, error) {
	return writeCronJob(cronCmd, false)
}

func RemoveCronJob(cronCmd string) (string, error) {
	return writeCronJob(cronCmd, true)
}

func writeCronJob(cronCmd string, remove bool) (string, error) {

	tempFile := "croncommand.txt"

	out, err := checkCrontabContent()
	if err != nil {
		return "", err
	}

	if remove && bytes.Contains(out.Bytes(), []byte(cronCmd)) {
		log.Println("found ")

		newCron := strings.ReplaceAll(out.String(), cronCmd, "")

		out.Reset()
		_, err = out.Write([]byte(newCron))
		if err != nil {
			return "", fmt.Errorf("failed to write remove cron: %v", err)
		}
	}

	f, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create cron file: %v", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}

		if err := os.Remove(tempFile); err != nil {
			log.Printf("failed to remove file: %v", err)
		}
	}()

	if !remove {
		_, err = out.Write([]byte(cronCmd + "\n"))
		if err != nil {
			return "", fmt.Errorf("failed to write bytes to file: %v", err)
		}
	}

	_, err = f.Write(out.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to write bytes to file: %v", err)
	}

	cmd := exec.Command("crontab", tempFile)
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run crontab cmd: %v", err)
	}

	crontabContent, err := checkCrontabContent()
	if err != nil {
		return "", err
	}

	return crontabContent.String(), nil
}

func checkCrontabContent() (bytes.Buffer, error) {
	cmd := exec.Command("crontab", "-l")

	var out bytes.Buffer
	cmd.Stdout = &out

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to pipe stderr: %v", err)
	}

	if err = cmd.Start(); err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to start crontab cmd: %v", err)
	}

	bs, err := ioutil.ReadAll(stderr)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to read stderr: %v", err)
	}

	if bytes.Contains(bs, []byte("no crontab")) {
		log.Println("empty crontab")
		return out, nil
	}

	if err := cmd.Wait(); err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to wait for crontab cmd: %v", err)
	}

	return out, nil

}
