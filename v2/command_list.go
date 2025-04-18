// Copyright 2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// List of command command of Command processing golang package.

package command

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"sort"
)

// Page item struct
type commandsListItem struct {
	Command   string `json:"command"`
	Params    string `json:"params"`
	Return    string `json:"return"`
	ProcessIn string `json:"processIn"`
	Descr     string `json:"descr"`
	Request   string `json:"request"`
	Response  string `json:"response"`
}

// AddCommandsList adds commands list command. The commands list command return
// list of all commands of this server in text format. The input function f
// should convert indata to map[string]string.
func (c *Commands) AddCommandsList(processIn ProcessIn, setFieldsets ...bool) {

	// Check setFieldset
	setFieldset := true
	if len(setFieldsets) > 0 {
		setFieldset = setFieldsets[0]
	}

	// handler converts input data to map[string]string and use it in
	// commandsListHandler
	handler := func(command *CommandData, processIn ProcessIn, indata any) (
		io.Reader, error) {

		vars, err := c.Vars(indata)
		if err != nil {
			return nil, err
		}
		return c.commandsHttpHandler(setFieldset, vars)
	}

	// handlerJson converts input data to map[string]string and use it in
	// commandsListHandler
	handlerJson := func(command *CommandData, processIn ProcessIn, indata any) (
		io.Reader, error) {

		return c.commandsJsonHandler()
	}

	c.Add("commjson", "Get json list of commands.", processIn,
		"", "json list of commands", "", "", handlerJson)
	if processIn&HTTP != 0 {
		returnDesc := "HTML list of commands"
		c.Add("commands", "Get html list of commands.", processIn,
			"", returnDesc, "", "", handler)
		if setFieldset {
			c.Add("commfilt", "Get html list of commands with filter.", processIn,
				"{http}/{webrtc}/{tru}/{ws}", returnDesc, "", "", handler)
		}
	}
}

// commandsJsonHandler returns array of commands in json format.
func (c *Commands) commandsJsonHandler() (io.Reader, error) {

	// Output array of commands
	var list []commandsListItem

	// Get list of commands
	for command, cmd := range c.Iter() {
		list = append(list, commandsListItem{
			command, cmd.Params, cmd.Return, cmd.ProcessIn.String(), cmd.Descr,
			cmd.Request, cmd.Response,
		})
	}

	// Sort output list
	sort.Slice(list, func(i, j int) bool {
		return list[i].Command < list[j].Command
	})

	data, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// commandsHttpHandler returns list of commands in html format.
func (c *Commands) commandsHttpHandler(setFieldset bool, vars map[string]string) (io.Reader, error) {

	var fieldset string

	if setFieldset {
		fieldset =
			`
	<fieldset>
	<legend>Choose processing in commands:</legend>
	<div>
		<input type="checkbox" id="http" name="http" onclick="onClickHandler()" checked />
		<label for="scales">Http</label>
	
		<input type="checkbox" id="webrtc" name="webrtc" onclick="onClickHandler()" checked />
		<label for="horns">Webrtc</label>
	
		<input type="checkbox" id="tru" name="tru" onclick="onClickHandler()" checked />
		<label for="horns">Tru</label>

		<input type="checkbox" id="websocket" name="websocket" onclick="onClickHandler()" checked />
		<label for="horns">Websocket</label>
	</div>
	</fieldset>
	<br/>
	`
	}

	t := `
	<!DOCTYPE html>
	<html lang="en">
	<body>
	<h1>Commands api</h1>
	` + fieldset + `
	<div>
		Number of commands: {{len .List}}
	</div>
	<br/>

	<script>
	function setValues() {
		const params = new URLSearchParams(document.location.search);

		const http = "{{.Filter.ProcessIn.Http}}";
		document.getElementById("http").checked = http=="true";

		const webrtc = "{{.Filter.ProcessIn.Webrtc}}";
		document.getElementById("webrtc").checked = webrtc=="true";

		const tru = "{{.Filter.ProcessIn.Tru}}";
		document.getElementById("tru").checked = tru=="true";

		const ws = "{{.Filter.ProcessIn.Websocket}}";
		document.getElementById("websocket").checked = ws=="true";
	}
	setValues();
	</script>

	<div class="list">
	{{range .List}}
		<div class="command">{{.Command}}</div>
		<div class="descr">{{.Descr}}</div>{{if .Params}}
		<div class="params">params: {{.Params}}</div>{{end}}
		<div class="params">return: {{.Return}}</div>{{end}}
		<div class="params">processing in: {{.ProcessIn}}</div>
		<br/>
	{{end}}
	</div>

	<script>
	function onClickHandler() {
		var chkHttp = document.getElementById("http").checked;
		var chkWebrtc = document.getElementById("webrtc").checked;
		var chkTru = document.getElementById("tru").checked;
		var chkWs = document.getElementById("websocket").checked;

		if (chkHttp && chkWebrtc && chkTru && chkWs) {
			window.location = '/commands';
			return;
		}
		window.location = '/commfilt/'+chkHttp+'/'+chkWebrtc+'/'+chkTru+'/'+chkWs;
	}
	</script>

	<style>
	.command {
		font-weight: bold;
	}
	.descr {
		font-size: small;
	}
	.list {
		max-width: 915px;
	}
	</style>
	</body>
	</html>`

	// Page struct
	type Page struct {
		List   []commandsListItem
		Filter struct {
			ProcessIn struct {
				Http      bool
				Webrtc    bool
				Tru       bool
				Websocket bool
			}
		}
	}

	// Template page data
	var page Page

	// Parse parameters
	page.Filter.ProcessIn.Http = vars["http"] != "false"
	page.Filter.ProcessIn.Webrtc = vars["webrtc"] != "false"
	page.Filter.ProcessIn.Tru = vars["tru"] != "false"
	page.Filter.ProcessIn.Websocket = vars["websocket"] != "false"

	// Get list of commands depending on filter
	for command, cmd := range c.Iter() {
		// Check processing filter
		if page.Filter.ProcessIn.Http && cmd.ProcessIn&HTTP != 0 ||
			page.Filter.ProcessIn.Webrtc && cmd.ProcessIn&WebRTC != 0 ||
			page.Filter.ProcessIn.Tru && cmd.ProcessIn&TRU != 0 ||
			page.Filter.ProcessIn.Websocket && cmd.ProcessIn&WS != 0 {

			page.List = append(page.List, commandsListItem{
				command, cmd.Params, cmd.Return, cmd.ProcessIn.String(), cmd.Descr,
				cmd.Request, cmd.Response,
			})
		}
	}

	// Sort page list
	sort.Slice(page.List, func(i, j int) bool {
		return page.List[i].Command < page.List[j].Command
	})

	// Execute template
	buf := new(bytes.Buffer)
	tmpl := template.Must(template.New("list").Parse(t))
	tmpl.Execute(buf, page)

	return buf, nil
}
