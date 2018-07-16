package main

import (
    "log"
    "net/http"
    "encoding/json"
    "mvdan.cc/xurls"
    "strings"
    "time"
    "bytes"
    "os"
)

type PrEvent struct {
    Action          string
    Pull_request    struct {
        Body            string
        Comments_url    string
    }
}

type UrlResult struct {
    Url    string
    Status bool
}

type GithubComment struct {
    Body    string    `json:"body"`
}


func urlVerifier(urlChan <-chan string, resultChan chan<- UrlResult) {
    var client = &http.Client{
        Timeout: time.Second * 10,
    }

    for url := range urlChan {
        resp, err := client.Get(url)
        if err == nil{
            defer resp.Body.Close()
        }
        resultChan <- UrlResult{url, err == nil && resp.StatusCode>=200 && resp.StatusCode <= 399}
    }
}

func checkUrls(text string, prResultUrl string){
    urlChan := make(chan string)
    resultChan := make(chan UrlResult)

    for w := 1; w <= 3; w++ {
        go urlVerifier(urlChan, resultChan)
    }

    urls := xurls.Relaxed().FindAllString(text, -1)
    for _, url := range urls{
        urlChan <- url
    }
    close(urlChan)

    var reachableUrls []string
    var unreachableUrls []string
    for i:= 0; i<len(urls); i++{
        result := <-resultChan
        if result.Status{
            reachableUrls = append(reachableUrls, result.Url)
        }else{
            unreachableUrls = append(unreachableUrls, result.Url)
        }
    }
    recordUrlResults(prResultUrl, reachableUrls, unreachableUrls)
}

func recordUrlResults(prResultUrl string, reachable []string, unreachable []string){
    body := "## URL reachability report"
    if len(reachable) > 0 {
        body += "\n### The following URLs are **reachable**:\n"
        body += strings.Join(reachable,"\n")
    }
    if len(unreachable) > 0 {
        body += "\n### The following URLs are **not reachable**:\n"
        body += strings.Join(unreachable,"\n")
    }
    if len(reachable) == 0 && len(unreachable) == 0{
        body += "\nNo URLs found in PR description."
    }
    jsonBody, err := json.Marshal(GithubComment{body})
    if err == nil{
        client := &http.Client{}
        req, _ := http.NewRequest("POST", prResultUrl, bytes.NewBuffer(jsonBody))
        req.Header.Add("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))
        resp, err := client.Do(req)
        defer resp.Body.Close()
        if err != nil{
            log.Println(resp)
        }
    }
}

func handler(resp http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var pr PrEvent
    err := decoder.Decode(&pr)
    if err != nil {
        panic(err) //todo friendlyfy
    }
    
    if pr.Action == "opened" || pr.Action == "edited" {
        go checkUrls(pr.Pull_request.Body, pr.Pull_request.Comments_url)
    } 
}


func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

