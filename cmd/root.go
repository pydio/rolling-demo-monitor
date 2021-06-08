package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	cells_sdk "github.com/pydio/cells-sdk-go/v2"
	"github.com/pydio/cells-sdk-go/v2/client"
	"github.com/pydio/cells-sdk-go/v2/client/meta_service"
	"github.com/pydio/cells-sdk-go/v2/client/user_service"
	"github.com/pydio/cells-sdk-go/v2/models"
	sdk_rest "github.com/pydio/cells-sdk-go/v2/transport/rest"
)

var (
	url        string
	user       string
	pwd        string
	skipVerify bool

	demoUsers = map[string]string{
		"admin": "admin",
		"alice": "alice",
	}
)

const (
	userAgent = "demo-monitor/1.1"
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Ping demo server",
	Long:  `This command sends a listUsers request to the demo server and then tries to list the workspaces for each of the default users`,
	Run: func(cmd *cobra.Command, args []string) {

		sanityCheck()

		sdkConfig := &cells_sdk.SdkConfig{
			Url:           url,
			User:          user,
			Password:      pwd,
			SkipVerify:    skipVerify,
			CustomHeaders: map[string]string{"User-Agent": userAgent},
		}

		ctx, t, e := sdk_rest.GetClientTransport(sdkConfig, false)
		if e != nil {
			log.Fatal(e)
		}
		apiClient := client.New(t, strfmt.Default)

		// list users
		param := &user_service.SearchUsersParams{
			Context: ctx,
		}

		result, err := apiClient.UserService.SearchUsers(param)
		if err != nil {
			fmt.Printf("could not list users: %s\n", err.Error())
			log.Fatal(err)
		}
		if len(result.Payload.Users) == 0 {
			er := fmt.Errorf("no user at all on this instance")
			log.Fatal(er)
		}
		var foundOne bool
		if len(result.Payload.Users) > 0 {

			users := ""
			for _, u := range result.Payload.Users {
				users += u.Login + ", "
			}
			fmt.Printf("Found %d users in this instance: %s.\n", len(result.Payload.Users), strings.TrimSuffix(users, ", "))
		}
		for u, p := range demoUsers {
			fmt.Println(" ----------------", u, "----------------")
			if e := listingUserFiles(u, p); e == nil {
				foundOne = true
			}
		}
		if !foundOne {
			log.Fatal("Could not find any workspace for any user: check the demo server to further investigate")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Cannot execute root command", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&url, "url", "a", "https://demo.pydio.com", "Full URL of the demo server")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "admin", "Admin login")
	rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "admin", "Admin password")
	rootCmd.PersistentFlags().BoolVar(&skipVerify, "skip-verify", false, "Skip TLS verification (unknown authority, uncorrect FQDN)...")
}

func listingUserFiles(currLogin, currPwd string) error {

	sdkConfig := &cells_sdk.SdkConfig{
		Url:           url,
		User:          currLogin,
		Password:      currPwd,
		SkipVerify:    skipVerify,
		CustomHeaders: map[string]string{"User-Agent": userAgent},
	}

	ctx, t, e := sdk_rest.GetClientTransport(sdkConfig, false)
	if e != nil {
		return fmt.Errorf("could not retrieve client transport for %s, cause: %s", currLogin, e.Error())
	}
	apiClient := client.New(t, strfmt.Default)

	params := &meta_service.GetBulkMetaParams{
		Body: &models.RestGetBulkMetaRequest{NodePaths: []string{
			"/*",
		}},
		Context: ctx,
	}

	result, err := apiClient.MetaService.GetBulkMeta(params)
	if err != nil {
		return fmt.Errorf("could not list meta %s", err.Error())
	}

	if len(result.Payload.Nodes) > 0 {
		fmt.Printf("* %d meta\n", len(result.Payload.Nodes))
		fmt.Println("USER ", currLogin)

		for _, u := range result.Payload.Nodes {
			fmt.Println("  -", u.Path)
		}
	}

	return nil
}

func sanityCheck() {

	msg := ""

	if url == "" {
		msg += "URL, "
	}
	if user == "" {
		msg += "user, "
	}
	if pwd == "" {
		msg += "password, "
	}

	if msg != "" {
		msg = strings.TrimSuffix(msg, ", ")
		log.Fatal("All flags are compuslory. Missings values for: " + msg)
	}
}
