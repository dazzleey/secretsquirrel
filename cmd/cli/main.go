package main

import (
	"database/sql"
	"fmt"
	"secretsquirrel/config"
	"secretsquirrel/database"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type DatabaseWithPath struct {
	path     string
	database *gorm.DB
}

var (
	cfg       config.Config
	databases []*DatabaseWithPath

	extraDBPaths []string

	rootCmd = &cobra.Command{
		Use:   "secretsqcli",
		Short: "A simple CLI for managing secretsquirrel",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	banCmd = &cobra.Command{
		Use:                   "ban [username | id] <reason>",
		Short:                 "Ban a user",
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		Run:                   banUser,
	}
	unbanCmd = &cobra.Command{
		Use:                   "unban [username | id]",
		Short:                 "Unban a user",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run:                   unbanUser,
	}
	setRankCmd = &cobra.Command{
		Use:                   "setrank [username | id] <mod | admin | user>",
		Short:                 "Set a user's rank",
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
		Run:                   setRank,
	}
)

func main() {
	cobra.OnInitialize(initCobra)

	rootCmd.PersistentFlags().StringSliceVarP(&extraDBPaths, "database", "d", []string{}, "additional database locations.")

	rootCmd.AddCommand(banCmd)
	rootCmd.AddCommand(unbanCmd)
	rootCmd.AddCommand(setRankCmd)
	rootCmd.Execute()
}

func initCobra() {
	config.LoadConfig(&cfg)
	initDB()
}

func initDB() {
	databases = append(databases,
		&DatabaseWithPath{
			path:     cfg.Bot.DatabasePath,
			database: database.InitDB(cfg.Bot.DatabasePath),
		},
	)

	for len(extraDBPaths) > 0 {
		for _, path := range extraDBPaths {
			databases = append(databases,
				&DatabaseWithPath{
					path:     path,
					database: database.InitDB(path),
				})
		}
	}

	if len(databases) == 0 {
		panic("No databases specified.")
	}
}

func setRank(cmd *cobra.Command, args []string) {
	var rank database.UserRank

	switch args[1] {
	case "admin":
		rank = database.RankAdmin
	case "mod":
		rank = database.RankMod
	case "user":
		rank = database.RankUser
	default:
		fmt.Println("invalid rank")
		return
	}

	for _, d := range databases {
		var (
			user *database.User
			err  error
		)

		fmt.Println(d.path)

		user, err = database.FindUser(d.database, database.ByUsernameOrID(args[0]))
		if err != nil {
			fmt.Println(err)
			return
		}

		d.database.Model(user).Update("rank", rank)
	}

	fmt.Printf("User: %v New Rank: %v", args[0], rank.String())
}

func banUser(cmd *cobra.Command, args []string) {

	reason := strings.Join(args[1:], " ")

	for _, d := range databases {
		var (
			user *database.User
			err  error
		)

		fmt.Println(d.path)

		user, err = database.FindUser(d.database, database.ByUsernameOrID(args[0]))
		if err != nil {
			fmt.Println(err)
			return
		}

		d.database.Model(user).Updates(database.User{
			Rank:            database.RankBanned,
			Left:            sql.NullTime{Time: time.Now(), Valid: true},
			BlacklistReason: reason,
		})
	}

	fmt.Printf("User banned.")
}

func unbanUser(cmd *cobra.Command, args []string) {

	for _, d := range databases {
		var (
			user *database.User
			err  error
		)

		fmt.Println(d.path)

		user, err = database.FindUser(d.database, database.ByUsernameOrID(args[0]))
		if err != nil {
			fmt.Println(err)
			return
		}

		d.database.Model(user).Updates(database.User{
			Rank:            database.RankUser,
			BlacklistReason: "",
		})
	}

	fmt.Printf("User unbanned.")
}
