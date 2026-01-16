# gostructui

![gostructui-example](https://github.com/user-attachments/assets/5c671fa4-e00a-46c7-b736-1119fcbdd037)

A Go Library of bubbletea models leveraging the `reflect` package to expose structs
as forms and menus directly to CLI users, allowing them to edit fields with primitive types.

## Motivation

I built `gostructui` just as soon as I realized that I needed an easy, all-in-one method to ask for
user form input through a CLI. We shouldn't always have to ask CLI users for _one thing at a time._
I personally like to expose config structs to CLI users so that they can set those values easily
through the CLI, and then I save the result.

## Usage

Right now, the only user-editable fields are:
- Strings
- Integers
- Booleans

The repo contains an example of how to use the package withn `./example/main.go`. Let's walk through it!

### Step 1: Establish the struct you wish to expose to the user.
In this struct, we build out a list of potential fields on a theoretical
job application to illustrate the idea.

**What To Know**:
- The `smname` tag establishes the title and formatting of the field. If the tag is not present,
the menu will fall back to the default name of the struct field itself. For example, you'll
see in the above demonstration that the `Email` field renders as we would expect despite the
lack of the `smname` tag.
- The `smdes` tag renders an optional description when the user hovers their cursor over the field.
- We'll discuss the `BlacklistedField` bit in a minute. It will illustrate another feature!
```go
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
```

### Step 2: Choose custom settings to apply to your menu, if you desire.
There are a number of custom settings we could apply to our menu, such as
changing the ibeam cursor rendered during string input, or what the field
cursor might look like.
Here, we write a custom header to render during form interaction.
Because we're using custom settings, we will have to initialize them
before setting any of the values on them.
*Never forget to do this! Zero values for menu settings are NOT the defaults.*
```go
	customMenuSettings := &gostructui.MenuSettings{}
	customMenuSettings.Init()
	customMenuSettings.Header = "Apply for this job: "
```

### Step 3: Provide a struct to use during menu input.
Of course, you'll need a struct to expose to your CLI users!
Here, we simply declare an empty one, but don't worry: if you need to provide a struct
with non-zero values, you can also do that! The bubbletea model will keep those values
intact, showing them to users as existing values within the field.
```go
	newApplication := applicationForm{}
```

### Step 4: Initialize a menu.
Provide a pointer to your struct, a list of fields used as a whitelist or blacklist, and any
custom settings. Hey, there's our `BlacklistedField` option we set earlier! See how our
argument passed to the `asBlacklist` parameter is set to `true`? It means that any fields
with the names given within the string slice to the left will be hidden from users. You can
see it in the demo above; the field doesn't show up!
```go
configEditMenu, err := gostructui.InitialTModelStructMenu(&newApplication, []string{"BlacklistedField"}, true, customMenuSettings)
	if err != nil {
		log.Fatal("Trouble generating the application.")
	}
```

### Step 5: Use the menu with the bubbletea package!
The menu is a bubbletea model! That is, it implements the bubbletea package!
We're now ready to run it through bubbletea and expose the menu to users to capture
their input! The result is the demo you saw above.
```go
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

			// newApplication: "Wow, I feel like a new struct!"
		}
		fmt.Println("Thank you for applying!")
		time.Sleep(time.Second * 5)
		os.Exit(0)
	}
```
You have now captured user input for one or more fields using the `gostructui` package!
Do what you need with these new values.
