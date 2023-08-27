package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunTag cmd for tag actions
func RunTag() *cobra.Command {
	var command = &cobra.Command{
		Use:   "tag",
		Short: "tag subcommand",
		Long:  `Do various tag actions`,
	}

	command.AddCommand(RunTagAdd())
	command.AddCommand(RunTagDelete())
	command.AddCommand(RunTagList())

	return command
}

// RunTagList cmd to list tags
func RunTagList() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "List tags",
		Long:  "List tags.",
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run tag list")
		return nil
	}

	return command
}

// RunTagAdd cmd to add tags
func RunTagAdd() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add tags",
		Long:  "Add tags",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a tag as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run tag add")
		return nil
	}

	return command
}

// RunTagDelete cmd to delete tags
func RunTagDelete() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:   "delete",
		Short: "Delete tags",
		Long:  "Delete tags.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a tag as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run tag delete")
		return nil
	}

	return command
}
