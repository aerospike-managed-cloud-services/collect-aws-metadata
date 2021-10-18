package main

import (
    "encoding/json"
    "errors"
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


// test `e`, write a message using the standard error format (and exit) if it is an error
func check(e error) {
    if e != nil {
        log.Fatalf("** %s: %s", MY_PROGRAM_NAME, e.Error())
    }
}


// uses log.Printf to write an info message using a standard format
func PrintInfo(msg string) {
    log.Println(MY_PROGRAM_NAME, ":", msg)
}


// create a textfile for Prometheus to read from the events, using the output argument (an open file)
func WriteMetrics(output *os.File, metadata *fetched_metadata, prefix string) (int, error) {
    lines := 0

    _, errFprintf := fmt.Fprintf(output,
        "aws_maintenance_event_count{instance=\"%s\"} %d\n",
        metadata.instanceID,
        len(metadata.events),
    )
    if errFprintf != nil {
        return lines, errFprintf
    }
    lines = lines + 1

    for _, ev := range metadata.events {
        evTime, errParse := time.Parse("02 Jan 2006 15:04:05 UTC", ev.NotBefore)
        if errParse != nil {
            return lines, errParse

        }
        _, errFprintf2 := fmt.Fprintf(output,
            "aws_maintenance_event{instance=\"%s\", code=\"%s\", id=\"%s\"} %d\n",
            metadata.instanceID,
            ev.Code,
            ev.EventId,
            evTime.Unix(),
            )
        if errFprintf2 != nil {
            return lines, errFprintf2
        }
        lines = lines + 1
    }
    return lines, nil
}


// fetch `url` with HTTP GET and return the body
func FetchURL(url string) ([]byte, error) {
    resp, errGet := http.Get(url)
    if errGet != nil {
        return nil, errGet
    }

    if resp.Body != nil {
        defer resp.Body.Close()
    }

    body, errRead := ioutil.ReadAll(resp.Body)
    if errRead != nil {
        return nil, errRead
    }

    return body, nil
}


// fetch instance metadata from the well-known AWS URLs and configured baseURL
func FetchMetadata(opt *collect_options) (*fetched_metadata, error) {
    ret := &fetched_metadata{}
    eventsURL := opt.baseURL + DEFAULT_SCHEDULED_PATH
    instanceURL := opt.baseURL + DEFAULT_INSTANCE_ID_PATH

    instance, errFetch1 := FetchURL(instanceURL)
    if errFetch1 != nil {
        return nil, errFetch1
    }

    ret.instanceID = string(instance[:])

    body, errFetch2 := FetchURL(eventsURL)
    if errFetch2 != nil {
        return nil, errFetch2
    }

    errJSON := json.Unmarshal(body, &ret.events)
    if errJSON != nil {
        return nil, errJSON
    }
    counted := fmt.Sprintf("Fetched %s; %d events", eventsURL, len(ret.events))
    PrintInfo(counted)

    return ret, nil
}


type collect_options struct {
    baseURL, metricPrefix, textfilesPath string
}


type maintenance_event struct {
    NotBefore string `json:"NotBefore"` //     "20 Jan 2019 09:00:43 GMT"
    Code string `json:"Code"` //     "system-reboot",
    Description string `json:"Description"` //     "scheduled reboot",
    EventId string `json:"EventId"` //     "instance-event-1d59937288b749b32",
    NotAfter string `json:"NotAfter"` //     "20 Jan 2019 09:17:23 GMT",
    State string `json:"State"` //     "active"
}


type fetched_metadata struct {
    instanceID string
    events []maintenance_event
}


func ParseArgs() (*collect_options, error) {
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
        baseURL: *baseURL,
        metricPrefix: *metricPrefix,
        textfilesPath: *textfilesPath,
    }

    if len(ret.textfilesPath) == 0 {
        return ret, errors.New("Required: --textfiles-path")
    }

    return ret, nil
}


func main() {
    opt, parseErr := ParseArgs()
    check(parseErr)

    fetchedMetadata, fetchErr := FetchMetadata(opt)
    check(fetchErr)

    now := time.Now().Format("20060102150405Z")
    pth := fmt.Sprintf("%s/collect-aws-metadata.%s.%d.prom", opt.textfilesPath, now, os.Getpid())
    openFile, errOpen := os.Create(pth)
    check(errOpen)

    _, errWrite := WriteMetrics(openFile, fetchedMetadata, opt.metricPrefix)
    check(errWrite)

    openFile.Close()

    okMessage := fmt.Sprintf("Wrote %s", pth)
    PrintInfo(okMessage)
}
