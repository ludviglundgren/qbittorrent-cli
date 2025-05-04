package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/ludviglundgren/qbittorrent-cli/internal/config"

	"github.com/autobrr/go-qbittorrent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// RunCategory cmd for category actions
func RunCategory() *cobra.Command {
	var command = &cobra.Command{
		Use:   "category",
		Short: "Category subcommand",
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

	var (
		output string
	)

	command.Flags().StringVar(&output, "output", "", "Print as [formatted text (default), json]")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		cats, err := qb.GetCategoriesCtx(ctx)
		if err != nil {
			return errors.Wrap(err, "could not get categories")
		}

		if len(cats) == 0 {
			log.Println("No categories found")
			return nil
		}

		switch output {
		case "json":
			res, err := json.Marshal(cats)
			if err != nil {
				return errors.Wrap(err, "could not marshal categories")
			}

			fmt.Println(string(res))

		default:
			if err := printCategoryList(cats); err != nil {
				return errors.Wrap(err, "could not print category list")
			}
		}

		return nil
	}

	return command
}

var categoryItemTemplate = `{{ range .}}
Name: {{.Name}}
Save path: {{.SavePath}}
{{end}}
`

func printCategoryList(categories map[string]qbittorrent.Category) error {
	tmpl, err := template.New("category-list").Parse(categoryItemTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(os.Stdout, categories)
	if err != nil {
		return errors.Wrap(err, "could not generate template")
	}

	return nil
}

// RunCategoryAdd cmd to add categories
func RunCategoryAdd() *cobra.Command {
	var (
		dry      bool
		savePath string
	)

	var command = &cobra.Command{
		Use:   "add",
		Short: "Add category",
		Long:  "Add new category",
		Example: `  qbt category add test-category
  qbt category add test-category --save-path "/home/user/torrents/test-category"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringVar(&savePath, "save-path", "", "Category default save-path. Optional. Defaults to dir in default save dir.")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		// args
		// first arg is path to torrent file
		category := args[0]

		if dry {
			log.Printf("dry-run: successfully added category: %s\n", category)

			return nil

		} else {
			if err := qb.CreateCategoryCtx(ctx, category, savePath); err != nil {
				return errors.Wrap(err, "could not create category")
			}

			log.Printf("successfully added category: %s\n", category)
		}

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
		Use:     "delete",
		Short:   "Delete category",
		Long:    "Delete category",
		Example: `  qbt category delete test-category`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		// args
		// first arg is path to torrent file
		category := args[0]

		if dry {
			log.Printf("dry-run: successfully deleted category: %s\n", category)

			return nil

		} else {
			if err := qb.RemoveCategoriesCtx(ctx, []string{category}); err != nil {
				return errors.Wrap(err, "could not delete category")
			}

			log.Printf("successfully deleted category: %s\n", category)
		}

		return nil
	}

	return command
}

// RunCategoryEdit cmd to edit category
func RunCategoryEdit() *cobra.Command {
	var (
		dry      bool
		savePath string
	)

	var command = &cobra.Command{
		Use:   "edit",
		Short: "Edit category",
		Long:  "Edit category",
		Example: `  qbt category edit test-category --save-path "/home/user/new/path"
  qbt category edit test-category --save-path ""`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a category as first argument")
			}

			return nil
		},
	}
	command.Flags().BoolVar(&dry, "dry-run", false, "Run without doing anything")
	command.Flags().StringVar(&savePath, "save-path", "", "Edit category save-path")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		config.InitConfig()

		qbtSettings := qbittorrent.Config{
			Host:      config.Qbit.Addr,
			Username:  config.Qbit.Login,
			Password:  config.Qbit.Password,
			BasicUser: config.Qbit.BasicUser,
			BasicPass: config.Qbit.BasicPass,
		}

		qb := qbittorrent.NewClient(qbtSettings)

		ctx := cmd.Context()

		if err := qb.LoginCtx(ctx); err != nil {
			return errors.Wrap(err, "could not login to qbit")
		}

		// args
		// first arg is path to torrent file
		category := args[0]

		if dry {
			log.Printf("dry-run: successfully edited category: %s\n", category)

			return nil

		} else {
			if err := qb.EditCategoryCtx(ctx, category, savePath); err != nil {
				return errors.Wrap(err, "could not edit category")
			}

			log.Printf("successfully edited category: %s\n", category)
		}

		return nil
	}

	return command
}
