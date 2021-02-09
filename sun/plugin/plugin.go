package plugin

type Plugin struct {
	//The name of the plugin it self
	Name string
	//The name of the authors of the plugin
	Authors []string
	//The api version this plugin should run on...
	Api string
	//The time in load order when this plugin should be loaded `STARTUP` or ``
	Load string
	//Path to main js file to be ran as a start up script almost
	Main string
	//A two option thing with `objective` amd `script`
	Type string
	//The version of the plugin
	Version string
}
