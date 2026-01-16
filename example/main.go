package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bntrtm/gostructui"
	tea "github.com/charmbracelet/bubbletea"
)

// STEP 1: Establish the struct you wish to expose to the user.

// applicationForm holds fields typical of a job application.
type applicationForm struct {
	FirstName        string `smname:"First Name"`
	LastName         string `smname:"Last Name"`
	Email            string
	PhoneNo          int    `smname:"Phone"`
	Country          string `smname:"Country"`
	Location         string `smname:"Location (City)"`
	CanTravel        bool   `smname:"Travel" smdes:"Can you travel for work?"`
	BlacklistedField string
}

func main() {
	// STEP 2: Choose custom settings to apply to your menu, if you desire.
	customMenuSettings := &gostructui.MenuSettings{}
	// Ensure that if you use custom settings, you initialize them first!
	customMenuSettings.Init()
	customMenuSettings.Header = "Apply for this job: "

	// STEP 3: Provide a struct to use.
	// Don't worry, if you need to provide a struct with non-zero
	// values, you can also do that! The tea model will keep those
	// values intact.
	newApplication := applicationForm{}
	// STEP 4: Initialize a menu!
	// Provide a pointer to your struct, blacklisted or
	// whitelisted fields, and any custom settings.
	configEditMenu, err := gostructui.InitialTModelStructMenu(&newApplication, []string{"BlacklistedField"}, true, customMenuSettings)
	if err != nil {
		log.Fatal("Trouble generating the application.")
	}
	// STEP 5: Use the menu---a bubbletea model---with the bubbletea package!
	// Here, we capture the result (our struct with user-entered values)
	// as the tea.Model variable "entry".
	p := tea.NewProgram(configEditMenu)
	if entry, err := p.Run(); err != nil {
		log.Fatal("Trouble generating the application.")
	} else {
		if entry.(gostructui.TModelStructMenu).QuitWithCancel {
			fmt.Printf("Canceled application.\n")
			os.Exit(0)
		} else {
			err = entry.(gostructui.TModelStructMenu).ParseStruct(&newApplication)
			if err != nil {
				log.Fatal("Trouble generating the application.")
			}

			// Your struct is now full of user-entered values!
			// Do what you need with it.

			// newApplication: "Wow, I feel like a new struct!"
		}
		fmt.Println("Thank you for applying!")
		time.Sleep(time.Second * 5)
		os.Exit(0)
	}
}
