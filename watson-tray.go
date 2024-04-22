package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/fsnotify/fsnotify"
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
// - Separator
// - Start new project: (open dialog asking for name and any defaul tags)
// - Separator
// - Quit

// Tooltip: Not running if not running otherwise first three lines of menu items

func onReady() {
	// var running_icon = mustLoadIcon("icons/running.png")
	// var not_running_icon = mustLoadIcon("icons/not_running.png")
	var elapsedTime time.Duration
	var byteValue []byte
	var jsonFile *os.File
	var err error
	var mProjectList []*systray.MenuItem

	mNotRunningIcon := mustLoadIcon("icons/not_running.png")
	mRunningIcon := mustLoadIcon("icons/running.png")

	var watsonState WatsonState
	// Read project state
	// TODO: Get rid of hardcoded values
	jsonStateFile, err = os.Open("/home/tmoyer/.config/watson/state")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonStateFile.Close()
	byteValue, _ = ioutil.ReadAll(jsonStateFile)

	json.Unmarshal(byteValue, &watsonState)

	systray.SetTooltip("Current status for watson time tracking")

	mDailyTime := systray.AddMenuItem("Total daily time", "Time accounted for today")
	mDailyTime.Enable()
	mDailyTime.Hide()
	
	systray.AddSeparator()

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
	mElapsedTime.Hide()

	systray.AddSeparator()

	// Project submenu
	mProjectSubmenu := systray.AddMenuItem("Existing projects", "Projects that watson already knows about")
	mProjectSubmenu.Enable()

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
		systray.SetTitle(fmt.Sprintf("%v (%v) [pending]", watsonState.Project, elapsedTime.String()))

		// Set icon
		systray.SetIcon(mRunningIcon)
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
		systray.SetIcon(mNotRunningIcon)
	}

	jsonFramesFile, err = os.Open("/home/tmoyer/.config/watson/frames")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFramesFile.Close()

	decoder := json.NewDecoder(jsonFramesFile)
	decoder.UseNumber()

	_, err = decoder.Token()
	if err != nil {
		log.Fatal(err)
	}
	definedProjects := map[string]bool{}

	for decoder.More() {
		// Decode one Frame
		var m [7]interface{}
		// decode an array value
		err := decoder.Decode(&m)
		if err != nil {
			log.Fatal(err)
		}

		definedProjects[m[2].(string)] = true
	}

	// read closing bracket
	_, err = decoder.Token()
	if err != nil {
		log.Fatal(err)
	}

	var found bool

	// Hide everything
	for _, projectMenuItem := range mProjectList {
		projectMenuItem.Hide()
	}
	
	for k, _ := range definedProjects {
		// fmt.Println("Project: ", k)
		// Does this menu item already exist?
		found = false
		for _, menuItem := range mProjectList {
			if menuItem.String() == k {
				found = true
				menuItem.Show()
			}
		}
		if found == false {
			newMenuItem := mProjectSubmenu.AddSubMenuItem(k, "")
			// fmt.Printf("%T : %+v\n", mProjectList, mProjectList)
			mProjectList = append(mProjectList, newMenuItem)
			newMenuItem.Show()
		}
	}

	// TODO: Start new project

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit watson-tray")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}() 

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				switch event.Name {
				case "/home/tmoyer/.config/watson/state", "/home/tmoyer/.config/watson/frames":
					fmt.Printf("EVENT! %#v\n", event)
					var watsonState WatsonState
					// Read project state
					// TODO: Get rid of hardcoded values
					jsonStateFile, err = os.Open("/home/tmoyer/.config/watson/state")

					// if we os.Open returns an error then handle it
					if err != nil {
						fmt.Println(err)
					}
					defer jsonStateFile.Close()
					byteValue, _ = ioutil.ReadAll(jsonStateFile)

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
						systray.SetIcon(mRunningIcon)
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
						systray.SetIcon(mNotRunningIcon)
					}
					// Read the list of projects
					// TODO: Get rid of hardcoded values
					jsonFramesFile, err := os.Open("/home/tmoyer/.config/watson/frames")

					// if we os.Open returns an error then handle it
					if err != nil {
						fmt.Println(err)
					}

					defer jsonFramesFile.Close()

					decoder := json.NewDecoder(jsonFramesFile)
					decoder.UseNumber()

					_, err = decoder.Token()
					if err != nil {
						log.Fatal(err)
					}
					definedProjects := map[string]bool{}

					for decoder.More() {
						// Decode one Frame
						var m [7]interface{}
						// decode an array value
						err := decoder.Decode(&m)
						if err != nil {
							log.Fatal(err)
						}

						definedProjects[m[2].(string)] = true
					}

					// read closing bracket
					_, err = decoder.Token()
					if err != nil {
						log.Fatal(err)
					}

					var found bool

					// Hide everything
					for _, projectMenuItem := range mProjectList {
						projectMenuItem.Hide()
					}

					for k, _ := range definedProjects {
						// fmt.Println("Project: ", k)
						// Does this menu item already exist?
						found = false
						for _, menuItem := range mProjectList {
							if menuItem.String() == k {
								found = true
								menuItem.Show()
							}
						}
						if found == false {
							newMenuItem := mProjectSubmenu.AddSubMenuItem(k, "")
							// fmt.Printf("%T : %+v\n", mProjectList, mProjectList)
							mProjectList = append(mProjectList, newMenuItem)
							newMenuItem.Show()
						}
					}
				default:
					fmt.Printf("Ignoring event %#v\n", event)
				}
			// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	if err := watcher.Add("/home/tmoyer/.config/watson/"); err != nil {
		fmt.Println("ERROR", err)
	}

	updateTicker := time.NewTicker(2 * time.Second)

	for range updateTicker.C {
		var watsonState WatsonState
		var daily_duration time.Duration
	    
		jsonFramesFile, err := os.Open("/home/tmoyer/.config/watson/frames")

		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}

		defer jsonFramesFile.Close()

		decoder := json.NewDecoder(jsonFramesFile)
	    decoder.UseNumber()

	    _, err = decoder.Token()
	    if err != nil {
	        log.Fatal(err)
	    }
	    
	    current_date := time.Now()
	    newYork, err := time.LoadLocation("America/New_York")
	    for decoder.More() {
	        // Decode one Frame
	        var m [7]interface{}
	        // decode an array value
	        err := decoder.Decode(&m)
	        if err != nil {
	            log.Fatal(err)
	        }

	        end_time_unix, err := m[1].(json.Number).Int64()
	        start_time_unix, err := m[0].(json.Number).Int64()

	        duration, err := time.ParseDuration(fmt.Sprintf("%vs", strconv.FormatInt(end_time_unix - start_time_unix, 10)))
	        start_time := time.Unix(start_time_unix, 0)

	        if current_date.Year() == start_time.Year() && current_date.YearDay() == start_time.YearDay() {
	            daily_duration += duration
	        }
	    }

		// Read project state
		jsonStateFile, err = os.Open("/home/tmoyer/.config/watson/state")

		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		defer jsonStateFile.Close()
		byteValue, _ = ioutil.ReadAll(jsonStateFile)

		json.Unmarshal(byteValue, &watsonState)

		if watsonState.Start != 0 {
			// Set and show elapsed time
			elapsedTime = time.Duration((time.Now().Unix() - watsonState.Start) * int64(time.Second))
			mElapsedTime.SetTitle(fmt.Sprintf("Elapsed time: %v", elapsedTime.String()))
			mElapsedTime.Show()

			// Set systray title
			systray.SetTitle(fmt.Sprintf("%v (%v) [%v]", watsonState.Project, elapsedTime.String()), daily_duration + elapsedTime)
		} else {
			// Set total tracked time
		}
	}

	// Ticker to update:
	// mElapsedTime
	// systray.Title
	// daily total

	// current_date := time.Now()
    // newYork, err := time.LoadLocation("America/New_York")
    // end_time_unix, err := m[1].(json.Number).Int64()
    // start_time_unix, err := m[0].(json.Number).Int64()

    // duration, err := time.ParseDuration(fmt.Sprintf("%vs", strconv.FormatInt(end_time_unix - start_time_unix, 10)))

    // start_time := time.Unix(start_time_unix, 0)

    // if current_date.Year() == start_time.Year() && current_date.YearDay() == start_time.YearDay() {
    //     fmt.Printf("We have a match between current and start (%v) == (%v)\n", start_time.Format("2006-01-02"), current_date.Format("2006-01-02"))
    //     daily_duration += duration
    // } else {
    //     fmt.Printf("We don't have a match between current and start (%v) =/= (%v)\n", start_time.Format("2006-01-02"), current_date.Format("2006-01-02"))
    // }

	<-done
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
