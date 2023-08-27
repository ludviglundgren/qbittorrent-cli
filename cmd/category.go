package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunCategory cmd for category actions
func RunCategory() *cobra.Command {
	var command = &cobra.Command{
		Use:   "category",
		Short: "category subcommand",
		Long:  `Do various category actions`,
	}

	command.AddCommand(RunCategoryAdd())
	command.AddCommand(RunCategoryDelete())
	command.AddCommand(RunCategoryEdit())
	command.AddCommand(RunCategoryList())

	return command
}

// RunCategoryList cmd to add categories
func RunCategoryList() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "List categories",
		Long:  `List categories`,
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("list categories")
		return nil
	}

	return command
}

// RunCategoryAdd cmd to add categories
func RunCategoryAdd() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add category",
		Long:  "Add new category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run category add")
		return nil
	}

	return command
}

// RunCategoryDelete cmd to delete categories
func RunCategoryDelete() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:   "delete",
		Short: "Delete category",
		Long:  "Delete category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run torrent delete")
		return nil
	}

	return command
}

// RunCategoryEdit cmd to edit category
func RunCategoryEdit() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:   "edit",
		Short: "Edit category",
		Long:  "Edit category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("run torrent edit")
		return nil
	}

	return command
}
