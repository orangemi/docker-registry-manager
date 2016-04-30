package controllers

import "github.com/astaxie/beego"

type ScheduleController struct {
	beego.Controller
}

// Get returns the template for the schedule action page
func (c *ScheduleController) Get() {

	// Index template
	c.TplName = "schedule.tpl"
}
