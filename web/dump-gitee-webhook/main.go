/*
 * 接收并保存 gitee webhook 事件发送的数据，包括：HTTP Query，HTTP Header，HTTP Body
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

const basePath = "dump"

type payload struct {
	Query  interface{} `json:"query"`
	Header interface{} `json:"header"`
	Body   interface{} `json:"body"`
}

func responseHTTPError(w http.ResponseWriter, response string, statusCode int) {
	log.Fatalln(response)
	http.Error(w, response, statusCode)
}

func save(filePath string, content []byte) error {
	return ioutil.WriteFile(path.Join(basePath, filePath), content, 0600)
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseHTTPError(w, "read body failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	var v interface{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		responseHTTPError(w, "couldn't unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// filename
	m := v.(map[string]interface{})
	hookName, _ := m["hook_name"].(string)
	action, _ := m["action"].(string)
	notableType, _ := m["noteable_type"].(string)
	filename := hookName
	switch hookName {
	case "push_hooks":
		filename += ".json"
	case "tag_push_hooks":
		filename += ".json"
	case "issue_hooks":
		filename += "-" + action + ".json"
	case "merge_request_hooks":
		filename += "-" + action + ".json"
	case "note_hooks":
		filename += "-" + action + "-" + notableType + ".json"
	}

	p := payload{
		Query:  r.URL.Query(),
		Header: r.Header,
		Body:   v,
	}
	data, _ := json.MarshalIndent(p, "", "    ")
	if save(filename, data) != nil {
		responseHTTPError(w, "save failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(string(data))

	fmt.Fprintln(w, "Event received.")
}

func main() {
	http.HandleFunc("/hook", handler)
	log.Fatalln(http.ListenAndServe(":1234", nil))
}
