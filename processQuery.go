// Copyright (C) 2013 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qslquery

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

    "code.google.com/p/google-api-go-client/mirror/v1"

	"appengine"
	"appengine/urlfetch"
)

func processQueries(r *http.Request, svc *mirror.Service) error {
	
	timelineWaiting, err := svc.Timeline.List().Do()
	if err != nil {
		return err
	}
	
	timelineItemsWaiting := timelineWaiting.Items
	var inReplyToText string
	for _, t := range timelineItemsWaiting {
		inReplyToText = t.InReplyTo
		if inReplyToText != "" {
			itemTextInput := t.Text
			itemWords := strings.Fields(itemTextInput)
			itemTextOutput := ""
			for _, word := range itemWords {
				addChars := ""
				if word != "" {
					switch strings.ToLower(word) {
						case "alfa","alpha" : addChars = "A"
						case "bravo","baker" : addChars = "B"
						case "charlie" : addChars = "C" 
						case "delta" : addChars = "D" 
						case "echo" : addChars = "E" 
						case "foxtrot","fox" : addChars = "F"
						case "trot" : addChars = "" 
						case "golf" : addChars = "G" 
						case "hotel" : addChars = "H" 
						case "india" : addChars = "I"
						case "juliet" : addChars = "J"
						case "kilo" : addChars = "K" 
						case "lima" : addChars = "L" 
						case "mike","mic" : addChars = "M" 
						case "november" : addChars = "N" 
						case "oscar","oskar" : addChars = "O" 
						case "papa" : addChars = "P" 
						case "quebec" : addChars = "Q" 
						case "romeo" : addChars = "R" 
						case "sierra" : addChars = "S"
						case "tango" : addChars = "T"
						case "uniform" : addChars = "U" 
						case "victor" : addChars = "V" 
						case "whiskey" : addChars = "W" 
						case "x-ray","xray","x" : addChars = "X"
						case "ray" : addChars = "" 
						case "yankee" : addChars = "Y" 
						case "zulu" : addChars = "Z"
						case "0","zero" : addChars = "0"
						case "1","one" : addChars = "1"
						case "2","two","to","too" : addChars = "2"
						case "3","three","tree" : addChars = "3"
						case "4","four" : addChars = "4"
						case "5","five","fife","fiver" : addChars = "5"
						case "6","six" : addChars = "6"
						case "7","seven" : addChars = "7"
						case "8","eight" : addChars = "8"
						case "9","nine","niner" : addChars = "9"			
						default : addChars = "ERR"
					}	
					itemTextOutput = itemTextOutput + addChars				
				}				
			}
			url := "http://callook.info/" + strings.ToUpper(itemTextOutput) + "/json"
		    c := appengine.NewContext(r)
		    client := urlfetch.Client(c)
		    res, err := client.Get(url)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if err != nil {
				return err
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			
			u := map[string]interface{}{}
			err = json.Unmarshal(body, &u)	
			if err != nil {
				return err
			}

			newText := ""
			if u["status"] == "VALID" {
				a := u["address"].(map[string]interface{})
				lic := u["current"].(map[string]interface{})
				newText = newText + "\n\"" + strings.ToUpper(itemTextInput) + "\""
				newText = newText + "\n" + strings.ToUpper(itemTextOutput)
				if u["name"].(string) != "" {
					newText = newText + ", " + u["name"].(string) 
				}
				if a["line1"].(string) != "" {
					newText = newText + ",\n" + a["line1"].(string)
				}
				if a["line2"].(string) != "" { 
					newText = newText + ", " + a["line2"].(string)
				}
				if a["attn"].(string) != "" { 
					newText = newText + ", " + a["attn"].(string)
				}
				if lic["operClass"].(string) != "" {
					newText = newText + "\nLicense: " + lic["operClass"].(string)
				}
				newText = newText + "\n\n#QSLquery"
				t.MenuItems = []*mirror.MenuItem{&mirror.MenuItem{Action: "SHARE"},&mirror.MenuItem{Action: "READ_ALOUD"},&mirror.MenuItem{Action: "TOGGLE_PINNED"},&mirror.MenuItem{Action: "DELETE"}}
			} else {
				newText = newText + "QSL Query error: Invalid call sign\n"
				newText = newText + "Submitted query: \"" + strings.ToUpper(itemTextInput) + "\""				
				t.MenuItems = []*mirror.MenuItem{&mirror.MenuItem{Action: "DELETE"},&mirror.MenuItem{Action: "READ_ALOUD"}}
			} 
			t.Text = newText
			t.InReplyTo = ""
			//svc.Timeline.Update(t.Id, t).Do()
			
			for _, o := range timelineItemsWaiting {
				//svc.Timeline.Update(o.Id, o).Do()
				if o.Id == inReplyToText {				
					svc.Timeline.Delete(o.Id).Do()
				}
			}
			insertCallsignSearchItem(r, svc)
			time.Sleep(5000 * time.Millisecond)	
			t.Notification = &mirror.NotificationConfig{Level: "DEFAULT"}
			svc.Timeline.Update(t.Id, t).Do()
			
		}
	}
	return nil
	
}