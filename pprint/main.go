package pprint

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

func Loading(title string, exec func() error) {
	err := spinner.New().Title(title).
		Action(func() {
			err := exec()
			if err != nil {
				fmt.Printf("unexpected error: %s \n", err.Error())
				os.Exit(1)
			}
		}).
		Run()

	if err != nil {
		os.Exit(130)
	}
}

func Form(title string, value *int, options ...huh.Option[int]) {
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(title).
				Options(
					options...,
				).
				Value(value),
		),
	).WithTheme(huh.ThemeDracula()).Run()
	if err != nil {
		os.Exit(130)
	}
}

func Confirm(title string, affirmative string, negative string, value *bool) {
	err := huh.NewConfirm().
		Title(title).
		Affirmative(affirmative).
		Negative(negative).
		Value(value).
		Run()
	if err != nil {
		os.Exit(130)
	}
}
