package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const DEFAULT_BASE_URL = "http://169.254.169.254"
const DEFAULT_SCHEDULED_PATH = "/latest/meta-data/events/maintenance/scheduled"
const DEFAULT_INSTANCE_ID_PATH = "/1.0/meta-data/instance-id"
const DEFAULT_TOKEN_PATH = "/latest/api/token"
const TOKEN_TTL = "21600" // 6 hours in seconds
const MY_PROGRAM_NAME = "collect-aws-metadata"

var VERSION string // to set this, build with --ldflags="-X main.VERSION=vx.y.z"

// make these replaceable in a test
var logFatalf func(format string, v ...interface{}) = log.Fatalf
var osExit func(code int) = os.Exit

type collect_options struct {
	baseURL, metricPrefix, textfilesPath string
	token                                string
}

type maintenance_event struct {
	NotBefore   string `json:"NotBefore"`   //     "20 Jan 2019 09:00:43 GMT"
	Code        string `json:"Code"`        //     "instance-reboot", "system-reboot", "system-maintenance", "instance-retirement", "instance-stop"
	Description string `json:"Description"` //     "scheduled reboot",
	EventId     string `json:"EventId"`     //     "instance-event-1d59937288b749b32",
	NotAfter    string `json:"NotAfter"`    //     "20 Jan 2019 09:17:23 GMT",
	State       string `json:"State"`       //     "active", "completed", "canceled"
}

type fetched_metadata struct {
	instanceID string
	events     []maintenance_event
}

type HTTPErrorStatusCode struct {
	url     string
	code    int
	message string
}

func (e *HTTPErrorStatusCode) Error() string {
	return fmt.Sprintf("<%s> %s", e.url, e.message)
}

var errMissingTextfilesPath = errors.New("required: --textfiles-path")
var errShowVersion = errors.New("(not an error) --version override")

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
func writeMetrics(writer io.Writer, metadata *fetched_metadata, prefix string) error {
	_, err := fmt.Fprintf(writer,
		"%saws_maintenance_event_count{cloud_instance=\"%s\"} %d\n",
		prefix,
		metadata.instanceID,
		len(metadata.events),
	)
	if err != nil {
		return err
	}

	for _, ev := range metadata.events {
		evTime, err := time.Parse("2 Jan 2006 15:04:05 GMT", ev.NotBefore)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(writer,
			"%saws_maintenance_event{cloud_instance=\"%s\", event_code=\"%s\", event_id=\"%s\", event_state=\"%s\", event_date=\"%s\", days_hence=\"%d\"} %d\n",
			prefix,
			metadata.instanceID,
			ev.Code,
			ev.EventId,
			ev.State,
			evTime.Format("Mon 2006/01/02"), // formatted date of event, with weekday
			int64(evTime.Sub(time.Now()).Hours()/24), // duration (in days) until event
			evTime.Unix(), // timestamp
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Function to fetch IMDSv2 token
func fetchToken(baseURL string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", baseURL+DEFAULT_TOKEN_PATH, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", TOKEN_TTL)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// If we get a 403 or 404, the instance might be using IMDSv1
	if resp.StatusCode == 403 || resp.StatusCode == 404 {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", &HTTPErrorStatusCode{url: baseURL + DEFAULT_TOKEN_PATH, code: resp.StatusCode, message: resp.Status}
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

// Update fetchURL to use token if available
func fetchURL(url string, token string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add token header if token is available
	if token != "" {
		req.Header.Set("X-aws-ec2-metadata-token", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, &HTTPErrorStatusCode{url: url, code: resp.StatusCode, message: resp.Status}
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, _ := io.ReadAll(resp.Body)
	return body, nil
}

// Update fetchMetadata to use token
func fetchMetadata(opt *collect_options) (*fetched_metadata, error) {
	ret := &fetched_metadata{}
	eventsURL := opt.baseURL + DEFAULT_SCHEDULED_PATH
	instanceURL := opt.baseURL + DEFAULT_INSTANCE_ID_PATH

	// Try to fetch token first (for IMDSv2)
	token, err := fetchToken(opt.baseURL)
	if err != nil {
		return nil, err
	}
	opt.token = token

	instance, err := fetchURL(instanceURL, opt.token)
	if err != nil {
		return nil, err
	}

	ret.instanceID = string(instance)

	body, err := fetchURL(eventsURL, opt.token)
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

func parseArgs(args []string) (*collect_options, error) {
	flagSet := flag.NewFlagSet(MY_PROGRAM_NAME, flag.ContinueOnError)

	var ret collect_options

	showVersion := flagSet.Bool(
		"version",
		false,
		"Show the version of the app",
	)
	flagSet.StringVar(
		&ret.baseURL,
		"base-url",
		DEFAULT_BASE_URL,
		"HTTP URL for the meta-data service (e.g. 'http://169.254.169.254')",
	)
	flagSet.StringVar(
		&ret.metricPrefix,
		"metric-prefix",
		"",
		"Prometheus metric names will be given this prefix",
	)
	flagSet.StringVar(
		&ret.textfilesPath,
		"textfiles-path",
		"",
		"(required) path to a directory of Prometheus metric textfiles, i.e. one being read by node_exporter",
	)
	flagSet.Parse(args)

	if *showVersion {
		return &ret, errShowVersion
	}

	if len(ret.textfilesPath) == 0 {
		return &ret, errMissingTextfilesPath
	}

	return &ret, nil
}

func main() {
	opt, err := parseArgs(os.Args[1:])
	if errors.Is(err, errShowVersion) {
		if VERSION == "" {
			fmt.Printf("%s %s\n", MY_PROGRAM_NAME, "undefined")
		} else {
			fmt.Printf("%s %s\n", MY_PROGRAM_NAME, VERSION) // notest
		}
		osExit(0)
		return // reachable in a test
	}
	check(err)

	fetchedMetadata, err := fetchMetadata(opt)
	check(err)

	created, err := os.Create(opt.textfilesPath + "/collect-aws-metadata.prom")
	check(err)
	fWriter := os.NewFile(created.Fd(), created.Name())

	err = writeMetrics(fWriter, fetchedMetadata, opt.metricPrefix)
	check(err)
	fWriter.Close()

	okMessage := fmt.Sprintf("Wrote %s", created.Name())
	printInfo(okMessage)
}
