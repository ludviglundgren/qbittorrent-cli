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

// RunTag cmd for tag actions
func RunTag() *cobra.Command {
	var command = &cobra.Command{
		Use:   "tag",
		Short: "Tag subcommand",
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

		tags, err := qb.GetTagsCtx(ctx)
		if err != nil {
			return errors.Wrap(err, "could not get tags")
		}

		if len(tags) == 0 {
			log.Println("No tags found")
			return nil
		}

		switch output {
		case "json":
			res, err := json.Marshal(tags)
			if err != nil {
				return errors.Wrap(err, "could not parshal tags")
			}

			fmt.Println(string(res))

		default:
			if err := printTagsList(tags); err != nil {
				return err
			}
		}

		return nil
	}

	return command
}

var tagItemTemplate = `{{ range .}}
Name: {{.}}
{{end}}
`

func printTagsList(tags []string) error {
	tmpl, err := template.New("tags-list").Parse(tagItemTemplate)
	if err != nil {
		return errors.Wrap(err, "could not create template")
	}

	err = tmpl.Execute(os.Stdout, tags)
	if err != nil {
		return errors.Wrap(err, "could not generate tags list template")
	}

	return nil
}

// RunTagAdd cmd to add tags
func RunTagAdd() *cobra.Command {
	var (
		dry bool
	)

	var command = &cobra.Command{
		Use:     "add",
		Short:   "Add tags",
		Long:    "Add tags",
		Example: `  qbt tag add tag1`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a tag as first argument")
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
		// first arg is tag
		tag := args[0]

		if dry {
			log.Printf("dry-run: successfully created tag: %s\n", tag)

			return nil

		} else {
			if err := qb.CreateTagsCtx(ctx, args); err != nil {
				return errors.Wrapf(err, "could not create tag: %s", tag)
			}

			log.Printf("successfully created tag: %s\n", tag)
		}
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
		Use:     "delete",
		Short:   "Delete tags",
		Long:    "Delete tags.",
		Example: `  qbt tag delete tag1`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a tag as first argument")
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
		// first arg is tag
		tag := args[0]

		if dry {
			log.Printf("dry-run: successfully deleted tag: %s\n", tag)

			return nil

		} else {
			if err := qb.DeleteTagsCtx(ctx, []string{tag}); err != nil {
				return errors.Wrapf(err, "could not delete tag: %s", tag)
			}

			log.Printf("successfully deleted tag: %s\n", tag)
		}
		return nil
	}

	return command
}
