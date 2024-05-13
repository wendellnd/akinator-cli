package cmd

import (
	aki "akinator/akinator"
	"akinator/pprint"
	trans "akinator/translation"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "akinator",
	Short: "I can find any character you're thinking of",
	Long:  `Akinator is a game that can guess any character you're thinking of.`,
	Run: func(cmd *cobra.Command, args []string) {
		language, _ := cmd.Flags().GetString("language")

		translation := trans.NewTranslation(language)

		akinator := aki.NewAkinator(language)

		pprint.Loading(translation.Translate("Starting Akinator..."), akinator.Start)

		rightGuess := false
		for !rightGuess {
			choice := 0
			pprint.Form(
				akinator.CurrentQuestion,
				&choice,
				huh.NewOption(akinator.Answers[0], 0),
				huh.NewOption(akinator.Answers[1], 1),
				huh.NewOption(akinator.Answers[2], 2),
				huh.NewOption(akinator.Answers[3], 3),
				huh.NewOption(akinator.Answers[4], 4),
			)

			var guess aki.AkinatorGuess
			pprint.Loading("", func() (err error) {
				guess, err = akinator.NextQuestion(choice)
				if err != nil {
					return err
				}
				return nil
			})

			foundGuess := guess.NameProposition != ""
			if foundGuess {
				pprint.Confirm(fmt.Sprintf("%s %s", translation.Translate("Akinator thinks it's"), guess.NameProposition),
					translation.Translate("Yes"),
					translation.Translate("No"),
					&rightGuess,
				)

				if !rightGuess {
					pprint.Loading("", akinator.KeepGuessing)
				}
			}
		}

		fmt.Println(translation.Translate("Thanks for playing!"))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("language", "l", trans.SupportedLanguages.GetDefault(), "set the akinator language")
}
