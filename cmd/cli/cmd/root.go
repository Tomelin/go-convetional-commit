/*
Copyright © 2024 Rafael Tomelin
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/Tomelin/go-convetional-commit/internal/entity"
	"github.com/Tomelin/go-convetional-commit/internal/service"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkGitCommand(); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive {
			interactive, err := interactiveMode()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			sc := service.NewServiceCommit()
			err = sc.Commit(interactive)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

		} else {
			nonInteractiveMode()
		}

		push, _ := cmd.Flags().GetBool("push")
		if push {
			sc := service.NewServiceCommit()
			err := sc.Push()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		reset, _ := cmd.Flags().GetUint("reset")
		if reset > 0 {

			sc := service.NewServiceCommit()

			err := sc.Reset(uint(reset))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		// commit, _ := cmd.Flags().GetString("commit")
		// if commit == "" {
		// 	fmt.Println("NON Commit message:", commit)
		// } else {
		// 	fmt.Println("Commit message:", commit)
		// }

		// commitType, _ := cmd.Flags().GetString("type")
		// if commitType == "" {
		// 	fmt.Println("NON Commit message:", commit)
		// } else {
		// 	fmt.Println("Commit message:", commit)
		// }
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "Interactive mode")
	rootCmd.Flags().BoolP("push", "p", false, "Push the commit")
	rootCmd.Flags().UintP("reset", "r", 0, "Number of head to reaset")
	rootCmd.Flags().StringP("subject", "s", "", "Subject message of commit")
	rootCmd.Flags().String("type", "", "Commit type (feat, fix, chore, docs, style, refactor, perf, test, ci, build, revert)")
	rootCmd.Flags().String("taskId", "", "Task id")
	rootCmd.Flags().String("body", "", "Body message of commit")
	rootCmd.Flags().String("emoji", "", "Put emoji in commit message")
}

// interactiveMode is a function that runs the interactive mode
// It uses the promptui library to create a prompt
// It returns an error if the prompt fails
// Exmaple of commit:
//
// feat: :sparkles: add new feature
//
// message in the body
// Resolve: #123
func interactiveMode() (*entity.Commit, error) {

	opts := entity.Commit{}

	// STARTS Commit type
	items := opts.Option.OptionsList()

	prompt := promptui.Select{
		Label:     "Select kind of commit",
		Items:     items,
		IsVimMode: false,
	}

	_, result, err := prompt.Run()
	if result == "" {
		return nil, errors.New("please select a commit type")
	}

	if err != nil {
		return nil, fmt.Errorf("select prompt failed %v", err)
	}
	opts.Option = opts.Option.FromString(result)
	opts.Choice = result
	// ENDS Commit type

	// STARTS Commit message
	validate := func(input string) error {
		if input == "" {
			return errors.New("commit message is required")
		}

		if len(input) > 40 {
			return errors.New("commit message must be at most 40 characters long")
		}

		matched, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9].{0,39}$`, input)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("commit message must start with a letter and be at most 40 characters long")
		}

		return nil
	}

	prompt2 := promptui.Prompt{
		Label:    "Write the commit message",
		Validate: validate,
	}

	commitMessage, err := prompt2.Run()
	if commitMessage == "" {
		return nil, errors.New("commit message is required")
	}

	if err != nil {
		return nil, fmt.Errorf("input prompt failed %v", err)
	}
	opts.Subject = commitMessage
	// ENDS Commit message

	// STARTS body message
	commitBody := promptui.Prompt{
		Label: "write a message in the commit body",
	}

	commitBodyResult, err := commitBody.Run()
	if err != nil {
		return nil, fmt.Errorf("input prompt failed %v", err)
	}
	opts.Comment = commitBodyResult
	// ENDS  body message

	// STARTS task status
	promptStatus := promptui.Select{
		Label:     "Select task status",
		Items:     opts.StatusType.StatusList(),
		IsVimMode: false,
	}

	_, result, err = promptStatus.Run()
	if err != nil {
		return nil, fmt.Errorf("select prompt failed %v", err)
	}
	opts.StatusType = opts.StatusType.FromString(result)
	// ENDS task status

	// STARTS task id
	taskID := promptui.Prompt{
		Label: "Write the task id",
	}

	taskIdMessage, err := taskID.Run()
	if err != nil {
		return nil, fmt.Errorf("input prompt failed %v", err)
	}
	opts.TaskID = taskIdMessage
	// ENDS task id

	// STARTS enable emoji
	emoji := promptui.Select{
		Label: "Enable emoji in commit message? (y/n)",
		Items: opts.EnableEmoji(),
	}

	_, enabledEmoji, err := emoji.Run()
	if err != nil {
		return nil, fmt.Errorf("input prompt failed %v", err)
	}

	if enabledEmoji == "true" {
		opts.Emoji = opts.Option.Emoji()
	}
	// ENDS enable emoji

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &opts, nil
}

func nonInteractiveMode() {
	fmt.Println("Non interactive mode")
}

func checkGitCommand() error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is not installed")
	}

	return nil
}