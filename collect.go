package main

import (
    "flag"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"
)


type maintenance_event struct {
    NotBefore string `json:"NotBefore"` //     "21 Jan 2019 09:00:43 GMT"
    Code string `json:"Code"` //     "system-reboot",
    Description string `json:"Description"` //     "scheduled reboot",
    EventId string `json:"EventId"` //     "instance-event-0d59937288b749b32",
    NotAfter string `json:"NotAfter"` //     "21 Jan 2019 09:17:23 GMT",
    State string `json:"State"` //     "active"
}


const DEFAULT_BASE_URL = "http://169.254.169.254"
const DEFAULT_SCHEDULED_PATH = "/1.0/latest/meta-data/events/maintenance/scheduled"
const DEFAULT_INSTANCE_ID_PATH = "/1.0/meta-data/instance-id"
const MY_PROGRAM_NAME = "collect-aws-metadata"


func check(e error) {
    if e != nil {
        PrintFatal(e.Error())
    }
}


// uses log.Fatalf to write a message using the standard error format (and exit)
func PrintFatal(msg string) {
    log.Fatalf("** %s: %s", MY_PROGRAM_NAME, msg)
}


// uses log.Printf to write an info message using a standard format
func PrintInfo(msg string) {
    log.Println(MY_PROGRAM_NAME, ":", msg)
}


// create a textfile for Prometheus to read from the events, using the output argument (an open file)
func WriteMetrics(output *os.File, events []maintenance_event, instance string, prefix string) {
    _, errFprintf := fmt.Fprintf(output,
        "aws_maintenance_event_count{instance=\"%s\"} %d\n",
        instance,
        len(events),
    )
    check(errFprintf)
    for _, ev := range events {
        evTime, errParse := time.Parse("02 Jan 2006 15:04:05 UTC", ev.NotBefore)
        check(errParse)
        fmt.Fprintf(output,
            "aws_maintenance_event{instance=\"%s\", code=\"%s\", id=\"%s\"} %d\n",
            instance,
            ev.Code,
            ev.EventId,
            evTime.Unix(),
            )
    }
}


// fetch `url` with HTTP GET and return the body; exit on errors
func Fetch(url string) []byte {
    resp, errGet := http.Get(url)
    check(errGet)

    if resp.Body != nil {
        defer resp.Body.Close()
    }

    body, errRead := ioutil.ReadAll(resp.Body)
    check(errRead)

    return body
}


func main() {
    baseURL := flag.String(
        "base-url",
        DEFAULT_BASE_URL,
        "HTTP URL for the meta-data service (e.g. 'http://169.254.169.254')",
    )
    metric_prefix := flag.String(
        "metric-prefix",
        "",
        "Prometheus metric names will be given this prefix",
    )
    textfiles_path := flag.String(
        "textfiles-path",
        "",
        "(required) path to a directory of Prometheus metric textfiles, i.e. one being read by node_exporter",
    )
    flag.Parse()
    if len(*textfiles_path) == 0 {
        PrintFatal("Required: --textfiles-path")
    }

    eventsURL := *baseURL + DEFAULT_SCHEDULED_PATH
    instanceURL := *baseURL + DEFAULT_INSTANCE_ID_PATH

    instance := string(Fetch(instanceURL)[:])

    body := Fetch(eventsURL)

    events := []maintenance_event{}
    errJSON := json.Unmarshal(body, &events)
    check(errJSON)

    counted := fmt.Sprintf("Fetched %s; %d events", eventsURL, len(events))
    PrintInfo(counted)

    now := time.Now().Format("20060102150405Z")
    pth := fmt.Sprintf("%s/collect-aws-metadata.%s.%d.prom", *textfiles_path, now, os.Getpid())
    openFile, errOpen := os.Create(pth)
    check(errOpen)

    WriteMetrics(openFile, events, instance, *metric_prefix)
    openFile.Close()

    okMessage := fmt.Sprintf("Wrote %s", pth)
    PrintInfo(okMessage)
}
