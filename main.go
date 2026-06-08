package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	coresdk "test/GoSdk"
)

func main() {
	ctx := context.Background()

	time.Sleep(2 * time.Second)
	coresdk.Log.Info(ctx, "golang gRPC SDK client started......")

	// 1. Get input parameters
	inputJSON, err := coresdk.Parameter.GetInputJSONString(ctx)
	if err != nil {
		coresdk.Log.Error(ctx, fmt.Sprintf("Failed to get input parameters: %v", err))
		return
	}
	coresdk.Log.Debug(ctx, fmt.Sprintf("Input parameters: %s", inputJSON))

	// 2. Read and log all input fields (demonstrating 11 editor types)
	var inputMap map[string]interface{}
	json.Unmarshal([]byte(inputJSON), &inputMap)

	urls, _ := inputMap["urls"].([]interface{})
	sources, _ := inputMap["sources"].([]interface{})
	searchTerms, _ := inputMap["searchTerms"].([]interface{})
	location, _ := inputMap["location"].(string)
	notes, _ := inputMap["notes"].(string)
	maxResults := 100
	if v, ok := inputMap["max_results"].(float64); ok {
		maxResults = int(v)
	}
	language, _ := inputMap["language"].(string)
	category := 1
	if v, ok := inputMap["category"].(float64); ok {
		category = int(v)
	}
	dataSections, _ := inputMap["data_sections"].([]interface{})
	skipClosed := false
	if v, ok := inputMap["skip_closed"].(bool); ok {
		skipClosed = v
	}
	sinceDate, _ := inputMap["since_date"].(string)

	coresdk.Log.Info(ctx, fmt.Sprintf("[requestList] urls: %v", urls))
	coresdk.Log.Info(ctx, fmt.Sprintf("[requestListSource] sources: %v", sources))
	coresdk.Log.Info(ctx, fmt.Sprintf("[stringList] searchTerms: %v", searchTerms))
	coresdk.Log.Info(ctx, fmt.Sprintf("[input] location: %s", location))
	coresdk.Log.Info(ctx, fmt.Sprintf("[textarea] notes: %s", notes))
	coresdk.Log.Info(ctx, fmt.Sprintf("[number] max_results: %d", maxResults))
	coresdk.Log.Info(ctx, fmt.Sprintf("[select] language: %s", language))
	coresdk.Log.Info(ctx, fmt.Sprintf("[radio] category: %d", category))
	coresdk.Log.Info(ctx, fmt.Sprintf("[checkbox] data_sections: %v", dataSections))
	coresdk.Log.Info(ctx, fmt.Sprintf("[switch] skip_closed: %v", skipClosed))
	coresdk.Log.Info(ctx, fmt.Sprintf("[datepicker] since_date: %s", sinceDate))

	// 3. Proxy configuration (read from environment variables)
	proxyDomain := os.Getenv("PROXY_DOMAIN")
	coresdk.Log.Info(ctx, fmt.Sprintf("Proxy domain: %s", proxyDomain))

	var proxyAuth string
	proxyAuth = os.Getenv("PROXY_AUTH")
	coresdk.Log.Info(ctx, fmt.Sprintf("Proxy authentication: %s", proxyAuth))

	// 4. Construct proxy URL
	var proxyURL string
	if proxyAuth != "" {
		proxyURL = fmt.Sprintf("socks5://%s@%s", proxyAuth, proxyDomain)
	}
	coresdk.Log.Info(ctx, fmt.Sprintf("Proxy URL: %s", proxyURL))

	// 5. Business logic - create HTTP client with proxy
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	if proxyURL != "" {
		proxyParsed, err := url.Parse(proxyURL)
		if err != nil {
			coresdk.Log.Error(ctx, fmt.Sprintf("Failed to parse proxy URL: %v", err))
			return
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyParsed),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		coresdk.Log.Info(ctx, "Proxy client configured")
	}

	// 6. Set table headers
	headers := []*coresdk.TableHeaderItem{
		{Label: "Primary Key", Key: "id", Format: "text"},
		{Label: "Title", Key: "title", Format: "text"},
		{Label: "Description", Key: "description", Format: "text"},
	}

	_, err = coresdk.Result.SetTableHeader(ctx, headers)
	if err != nil {
		coresdk.Log.Error(ctx, fmt.Sprintf("Set table header failed: %v", err))
		return
	}

	// 7. Push data in batches (limited by maxResults)
	batchSize := 100
	sleepSeconds := 1

	for index := 1; index <= maxResults; index++ {
		data := map[string]any{
			"id":          fmt.Sprintf("test-%d", index),
			"title":       fmt.Sprintf("Test Title %d", index),
			"description": fmt.Sprintf("This is test description number %d", index),
		}

		_, err = coresdk.Result.UpsertData(ctx, data, "id")
		if err != nil {
			coresdk.Log.Error(ctx, fmt.Sprintf("Upsert data failed: %v", err))
			return
		}

		if index%batchSize == 0 {
			coresdk.Log.Info(ctx, fmt.Sprintf("Pushed %d items", index))
			if index < maxResults {
				time.Sleep(time.Duration(sleepSeconds) * time.Second)
			}
		}
	}

	coresdk.Log.Info(ctx, "Starting second push for multiples of 3")
	for index := 3; index <= maxResults; index += 3 {
		data := map[string]any{
			"id":          fmt.Sprintf("test-%d", index),
			"title":       fmt.Sprintf("Test Title %d", index),
			"description": fmt.Sprintf("This is updated test description number %d after second push", index),
		}

		_, err = coresdk.Result.UpsertData(ctx, data, "id")
		if err != nil {
			coresdk.Log.Error(ctx, fmt.Sprintf("Upsert data failed: %v", err))
			return
		}
	}

	coresdk.Log.Info(ctx, "Second push for multiples of 3 completed")
	coresdk.Log.Info(ctx, "Script execution completed")
}
