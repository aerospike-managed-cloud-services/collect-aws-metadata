package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const DEFAULT_BASE_URL = "http://169.254.169.254"
const DEFAULT_SCHEDULED_PATH = "/1.0/latest/meta-data/events/maintenance/scheduled"
const DEFAULT_INSTANCE_ID_PATH = "/1.0/meta-data/instance-id"
const MY_PROGRAM_NAME = "collect-aws-metadata"

// make Fatalf replaceable in a test
var logFatalf = log.Fatalf

// test `e`, write a message using the standard error format (and exit) if it is an error
func check(e error) {
	if e != nil {
		logFatalf("** %s: %s", MY_PROGRAM_NAME, e.Error())
	}
}

// uses log.Println to write an info message using a standard format
func printInfo(msg string) string {
	ret := MY_PROGRAM_NAME + ": " + msg
	log.Println(ret)
	return ret
}

// create a textfile for Prometheus to read from the events, using the output argument (an open file)
func writeMetrics(output *os.File, metadata *fetched_metadata, prefix string) error {
	_, err := fmt.Fprintf(output,
		"aws_maintenance_event_count{instance=\"%s\"} %d\n",
		metadata.instanceID,
		len(metadata.events),
	)
	if err != nil {
		return err
	}

	for _, ev := range metadata.events {
		evTime, err := time.Parse("02 Jan 2006 15:04:05 UTC", ev.NotBefore)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output,
			"aws_maintenance_event{instance=\"%s\", code=\"%s\", id=\"%s\"} %d\n",
			metadata.instanceID,
			ev.Code,
			ev.EventId,
			evTime.Unix(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// fetch `url` with HTTP GET and return the body
func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// fetch instance metadata from the well-known AWS URLs and configured baseURL
func fetchMetadata(opt *collect_options) (*fetched_metadata, error) {
	ret := &fetched_metadata{}
	eventsURL := opt.baseURL + DEFAULT_SCHEDULED_PATH
	instanceURL := opt.baseURL + DEFAULT_INSTANCE_ID_PATH

	instance, err := fetchURL(instanceURL)
	if err != nil {
		return nil, err
	}

	ret.instanceID = string(instance)

	body, err := fetchURL(eventsURL)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &ret.events)
	if err != nil {
		return nil, err
	}
	counted := fmt.Sprintf("Fetched %s; %d events", eventsURL, len(ret.events))
	printInfo(counted)

	return ret, nil
}

type collect_options struct {
	baseURL, metricPrefix, textfilesPath string
}

type maintenance_event struct {
	NotBefore   string `json:"NotBefore"`   //     "20 Jan 2019 09:00:43 GMT"
	Code        string `json:"Code"`        //     "system-reboot",
	Description string `json:"Description"` //     "scheduled reboot",
	EventId     string `json:"EventId"`     //     "instance-event-1d59937288b749b32",
	NotAfter    string `json:"NotAfter"`    //     "20 Jan 2019 09:17:23 GMT",
	State       string `json:"State"`       //     "active"
}

type fetched_metadata struct {
	instanceID string
	events     []maintenance_event
}

type FlagError struct {
	message string
}

func (e *FlagError) Error() string {
	return e.message
}

func parseArgs() (*collect_options, error) {
	baseURL := flag.String(
		"base-url",
		DEFAULT_BASE_URL,
		"HTTP URL for the meta-data service (e.g. 'http://169.254.169.254')",
	)
	metricPrefix := flag.String(
		"metric-prefix",
		"",
		"Prometheus metric names will be given this prefix",
	)
	textfilesPath := flag.String(
		"textfiles-path",
		"",
		"(required) path to a directory of Prometheus metric textfiles, i.e. one being read by node_exporter",
	)
	flag.Parse()

	ret := &collect_options{
		baseURL:       *baseURL,
		metricPrefix:  *metricPrefix,
		textfilesPath: *textfilesPath,
	}

	if len(ret.textfilesPath) == 0 {
		return ret, &FlagError{"Required: --textfiles-path"}
	}

	return ret, nil
}

func main() {
	opt, err := parseArgs()
	check(err)

	fetchedMetadata, err := fetchMetadata(opt)
	check(err)

	pth := fmt.Sprintf("%s/collect-aws-metadata.prom", opt.textfilesPath)
	openFile, err := os.Create(pth)
	check(err)

	err = writeMetrics(openFile, fetchedMetadata, opt.metricPrefix)
	check(err)

	openFile.Close()

	okMessage := fmt.Sprintf("Wrote %s", pth)
	printInfo(okMessage)
}
