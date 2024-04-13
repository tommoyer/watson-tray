package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/getlantern/systray"
)

type WatsonState struct {
	// defining struct variables
	Project string
	Start   int64
	Tags    []string
	Note    string
}

// Menu Items:
// - Current project: ("Not running" if not running)
// - Start time: (hide if not running)
// - Elapsed time: (hide if not running)
// - Projects submenu:
//   + List of known projects (click to start) // Must parse the config to grab tags to build the right state
// - Start new project: (open dialog asking for name and any defaul tags)
// - Separator
// - Quit

// Tooltip: Not running if not running otherwise first three lines of menu items

func onReady() {
	var running_icon = mustLoadIcon("icons/running.png")
	var not_running_icon = mustLoadIcon("icons/not_running.png")
	var elapsedTime time.Duration
	var err error
	var jsonFile *os.File
	var byteValue []byte

	systray.SetIcon(not_running_icon)
	systray.SetTitle("Watson Status")
	systray.SetTooltip("Current status for watson time tracking")

	// Add project menu time
	mProject := systray.AddMenuItem("Not running", "Current project")
	mProject.Enable()

	// Start time
	mStartTime := systray.AddMenuItem("Start time", "Time current work was started")
	mStartTime.Enable()
	mStartTime.Hide()

	// Elapsed time
	mElapsedTime := systray.AddMenuItem("Elapsed time", "How long has this task been worked on")
	mElapsedTime.Enable()
	mStartTime.Hide()

	systray.AddSeparator()

	// Project submenu
	mProjectSubmenu := systray.AddMenuItem("Existing projects", "Projects that watson already knows about")
	mProjectSubmenu.Enable()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit watson-tray")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	updateTicker := time.NewTicker(2 * time.Second)
	var count = 0
	for range updateTicker.C {
		var watsonState WatsonState
		// Read project state
		jsonFile, err = os.Open("/home/tmoyer/.config/watson/state")

		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		defer jsonFile.Close()
		byteValue, _ = ioutil.ReadAll(jsonFile)

		json.Unmarshal(byteValue, &watsonState)

		if watsonState.Start != 0 {
			// Set and show project name
			mProject.SetTitle(fmt.Sprintf("Project: %v", watsonState.Project))

			// Set and show start time
			mStartTime.SetTitle(fmt.Sprintf("Start time: %v", time.Unix(watsonState.Start, 0).Format(time.RFC822)))
			mStartTime.Show()

			// Set and show elapsed time
			elapsedTime = time.Duration((time.Now().Unix() - watsonState.Start) * int64(time.Second))
			mElapsedTime.SetTitle(fmt.Sprintf("Elapsed time: %v", elapsedTime.String()))
			mElapsedTime.Show()

			// Set systray title
			systray.SetTitle(fmt.Sprintf("%v (%v)", watsonState.Project, elapsedTime.String()))

			// Set icon
			systray.SetIcon(running_icon)
		} else {
			// Set project to "Not running"
			mProject.SetTitle("Not running")

			// Hide start time
			mStartTime.Hide()

			// Hide elapsed time
			mElapsedTime.Hide()

			// Set systray title
			systray.SetTitle("Not running")

			// Set icon
			systray.SetIcon(not_running_icon)
		}
		count++
		if count == 30 {
			count = 0
		}
	}
}

func onExit() {
	// cleanup here
}

func mustLoadIcon(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	systray.Run(onReady, onExit)
}
