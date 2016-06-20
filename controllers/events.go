package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/stefannaglee/docker-registry-manager/models/registry"
)

type EventsController struct {
	beego.Controller
}

// PostEvents accepts events sent from registries
func (c *EventsController) PostEvents() {

	es := registry.EventData{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &es); err != nil {
		c.CustomAbort(404, "Invalid body")
	}

	for _, event := range es.Events {
		registry.ActiveEvents[event.ID] = event
	}

	fmt.Println(registry.ActiveEvents)

	// Index template
	c.CustomAbort(200, "Success")
}
