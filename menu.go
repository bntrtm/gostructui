// Package gostructui provides bubbletea models that make it easy to
// expose forms and menus directly to CLI users.
package gostructui

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuSettings struct {
	NavCursorChar  string // cursor during navigation
	EditCursorChar string // cursor during edit
	IBeamChar      string // character shown right of text during edit
	TabAfterEntry  bool   // whether or not to jump to the next field after tabAfterEntry
	Header         string // message to display above the struct menu
}

type FieldKind int

const (
	FieldString FieldKind = iota
	FieldBool
	FieldInt
)

type menuField struct {
	kind FieldKind // value assigned to field
	s    string    // possible string value
	b    bool      // possible bool value
	i    int       // possible int value

	editBuf string // buffer for editing this field
	errBuf  string // potential error from bad input

	name   string // name of the struct field
	smName string // description pulled from smname tag
	smDes  string // description pulled from smdes tag
}

func (f *menuField) handleChar(char string) {
	switch f.kind {
	case FieldInt:
		if (char >= "0" && char <= "9") || (char == "-" && len(f.editBuf) == 0) {
			f.editBuf += string(char)
		}
	case FieldString:
		f.editBuf += string(char)
	case FieldBool:
		switch char {
		case "t", "1":
			f.b = true
		case "f", "0":
			f.b = false
		case "right", "left":
			f.b = !f.b
		}
	}
}

func (f *menuField) handleBackspace() {
	if len(f.editBuf) == 0 {
		return
	}
	f.editBuf = f.editBuf[:len(f.editBuf)-1]
}

func (f *menuField) render(editing bool, iBeamChar string) string {
	switch f.kind {
	case FieldInt:
		if editing {
			return f.editBuf + iBeamChar
		}
		return strconv.Itoa(f.i)
	case FieldString:
		if editing {
			return f.editBuf + iBeamChar
		}
		return f.s
	case FieldBool:
		if editing {
			if f.b {
				return "[t] ||  f "
			}
			return " t  || [f]"
		}
		return fmt.Sprintf("%v", f.b)
	default:
		return ""
	}
}

func (f *menuField) commitEdit() {
	switch f.kind {
	case FieldInt:
		if f.editBuf == "" || f.editBuf == "-" {
			f.i = 0
			return
		}
		v, err := strconv.Atoi(f.editBuf)
		if err != nil {
			f.errBuf = err.Error()
			return
		}
		f.i = v
	case FieldString:
		f.s = f.editBuf
	}

	f.editBuf = ""
	f.errBuf = ""
}

// getFieldName returns a name for the menu field.
// If an override name was provided via the smname tag
// (e.g. for human readability or foramtting), that will
// be returned. Otherwise, the name of the struct field
// is returned.
func (f *menuField) getFieldName() string {
	if f.smName != "" {
		return f.smName
	}
	return f.name
}

// TModelStructMenu is a bubbletea model that can be used to expose
// primitive struct fields to end users for input,
// as if they were elements of a menu.
type TModelStructMenu struct {
	// MENU STATE
	// fields which can be edited; populated dynamically
	menuFields     []menuField
	cursor         int  // which field our cursor is pointing at
	isEditingValue bool // tracks state of field editing
	QuitWithCancel bool // can be used to communicate whether changes ought be saved
	Settings       MenuSettings
}

// Init initializes the menu settings with default values.
// When using custom settings, this should be called first,
// before then overriding specific default values with
// those desired.
func (m *MenuSettings) Init() {
	*m = MenuSettings{
		IBeamChar:      "|",
		NavCursorChar:  "> ",
		EditCursorChar: ">>",
		TabAfterEntry:  true,
	}
}

// incrCursor increases the field index the user is focused on
func (m *TModelStructMenu) incrCursor() {
	if m.cursor > 0 {
		m.getFieldUnderCursor().errBuf = ""
		m.cursor--
	}
}

// decrCursor decreases the field index the user is focused on
func (m *TModelStructMenu) decrCursor() {
	m.getFieldUnderCursor().errBuf = ""
	if m.cursor < len(m.menuFields)-1 {
		m.cursor++
	}
}

func (m *TModelStructMenu) getFieldAtIndex(i int) *menuField {
	return &m.menuFields[i]
}

func (m *TModelStructMenu) getFieldUnderCursor() *menuField {
	return m.getFieldAtIndex(m.cursor)
}

// InitialTModelStructMenu creates a new struct menu from the given parameters.
// If customSettings are not provided, the menu will fall back to defaults.
// If using custom menu settings, first initialize them with the setDefaults() method.
func InitialTModelStructMenu(structObj any, fieldList []string, asBlacklist bool, customSettings *MenuSettings) (TModelStructMenu, error) {
	// if fieldList is empty, all fields are exposed to users; otherwise, it is used as a whitelist.
	// if bool parameter 'asBlacklist' is 'true', the fieldList is used as a blacklist instead of a whitelist.
	t := reflect.TypeOf(structObj)
	v := reflect.ValueOf(structObj)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	} else {
		return TModelStructMenu{}, errors.New("structObj should be a pointer to struct, so as to have addressable fields")
	}
	if t.Kind() != reflect.Struct {
		fmt.Println("ERROR: Not a struct. Check your input!")
		return TModelStructMenu{}, nil
	}
	newModel := TModelStructMenu{
		isEditingValue: false,
		menuFields:     []menuField{},
		QuitWithCancel: false,
	}

	if customSettings != nil {
		newModel.Settings = *customSettings
	} else {
		newModel.Settings.Init()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if len(fieldList) != 0 {
			if asBlacklist {
				if slices.Contains(fieldList, field.Name) {
					continue
				}
			} else {
				if !(slices.Contains(fieldList, field.Name)) {
					continue
				}
			}
		}

		fieldVal := v.FieldByName(field.Name)
		if !fieldVal.CanSet() {
			fmt.Printf("Warning: Field '%s' left unexposed (cannot be set; unexported or not addressable).\n", field.Name)
			continue
		}

		newField := menuField{}
		switch field.Type.Kind() {
		case reflect.String:
			newField.kind = FieldString
			newField.s = fieldVal.String()
		case reflect.Bool:
			newField.kind = FieldBool
			newField.b = fieldVal.Bool()
		case reflect.Int:
			newField.kind = FieldInt
			newField.i = int(fieldVal.Int())
		default:
			return TModelStructMenu{}, fmt.Errorf("could not parse struct")
		}
		newField.name = field.Name
		newField.smName = field.Tag.Get("smname")
		newField.smDes = field.Tag.Get("smdes")
		newModel.menuFields = append(newModel.menuFields, newField)
	}

	if len(newModel.menuFields) == 0 {
		return TModelStructMenu{}, fmt.Errorf("ERROR: No fields to expose to users in struct")
	}

	return newModel, nil
}

func (m TModelStructMenu) ParseStruct(obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %v", v.Kind())
	}
	v = v.Elem()

	for _, f := range m.menuFields {
		field := v.FieldByName(f.name)

		if !field.IsValid() {
			fmt.Printf("Warning: Field '%s' not found in struct.\n", f.name)
			continue
		}
		if !field.CanSet() {
			fmt.Printf("Warning: Field '%s' cannot be set (unexported or not addressable).\n", f.name)
			continue
		}

		switch f.kind {
		case FieldString:
			field.SetString(f.s)
		case FieldBool:
			field.SetBool(f.b)
		case FieldInt:
			field.SetInt(int64(f.i))
		default:
			return fmt.Errorf("unsupported kind for field '%s': %v", f.name, f.kind)
		}
	}

	return nil
}

func (m TModelStructMenu) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m TModelStructMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// toggle edit mode on field if 'enter' key was pressed
		if msg.String() == "enter" {
			f := m.getFieldUnderCursor()
			if !m.isEditingValue {
				m.isEditingValue = true
			} else {
				f.commitEdit()
				m.isEditingValue = false
				if m.Settings.TabAfterEntry {
					m.decrCursor()
				}
			}
		} else if msg.Type == tea.KeyBackspace {
			if m.isEditingValue {
				m.getFieldUnderCursor().handleBackspace()
			}
		} else {
			if m.isEditingValue {
				m.getFieldUnderCursor().handleChar(msg.String())
			} else {
				// Cool, what was the actual key pressed?
				switch msg.String() {

				case "s":
					return m, tea.Quit

				// These keys should exit the program.
				case "ctrl+c", "q":
					m.QuitWithCancel = true
					return m, tea.Quit

				// The "up" and "k" keys move the cursor up, or users may tab backward.
				case "up", "k", "shift+tab":
					m.incrCursor()

				// The "down" and "j" keys move the cursor down, or users may tab forward.
				case "down", "j", "tab":
					m.decrCursor()

				}
			}
		}
	}

	// Return the updated TModelStructMenu to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m TModelStructMenu) View() string {
	var s string
	// Add the header, if it exists
	if m.Settings.Header != "" {
		s = m.Settings.Header + "\n"
	}
	s += "\n"

	// for formatting, get longest field name
	maxFieldName := 0
	for _, field := range m.menuFields {
		if fieldName := field.getFieldName(); len(fieldName) > maxFieldName {
			maxFieldName = len(fieldName)
		}
	}

	// for formatting, get longest cursor string and build
	// the empty version of the cursor based on its length
	cursorEmpty := ""
	for _, cursor := range []string{m.Settings.NavCursorChar, m.Settings.EditCursorChar} {
		if len(cursor) > len(cursorEmpty) {
			cursorEmpty = ""
			for range cursor {
				cursorEmpty += " "
			}
		}
	}

	// Iterate over our fields
	for i, f := range m.menuFields {

		// Is the cursor pointing at this choice?
		cursor := "  " // no cursor
		if m.cursor == i {
			if m.isEditingValue {
				cursor = m.Settings.EditCursorChar
			} else {
				cursor = m.Settings.NavCursorChar
			}
		}

		// string represenation of field value
		value := f.render(m.isEditingValue && m.cursor == i, m.Settings.IBeamChar)
		s += fmt.Sprintf("%s ⟦ %-*s ⟧: %s\n", cursor, maxFieldName, f.getFieldName(), value)
	}

	// The footer
	s += "\n"
	if smDes := m.getFieldAtIndex(m.cursor).smDes; smDes != "" {
		s += smDes
	}
	s += "\n"

	s += "\nPress s to save and quit.\nPress q to quit without saving.\n"
	if f := m.getFieldUnderCursor(); f.errBuf != "" {
		s += fmt.Sprintf("ERROR: %s\n", f.errBuf)
	}

	// Send the UI for rendering
	return s
}
