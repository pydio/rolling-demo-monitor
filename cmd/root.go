package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/go-openapi/strfmt"

	"github.com/pydio/cells-sdk-go/client/meta_service"
	"github.com/pydio/cells-sdk-go/client/user_service"
	"github.com/pydio/cells-sdk-go/models"
	"github.com/pydio/cells-sdk-go"
	"github.com/pydio/cells-sdk-go/transport"
	"github.com/pydio/cells-sdk-go/client"
)

var (
	protocol   string
	host       string
	id         string
	user       string
	pwd        string
	skipVerify bool
	secret     string

	knownPwd = map[string]string{
		"admin": "admin",
		"demo":  "demo",
		"bob":   "P@ssw0rd",
		"alice": "P@ssw0rd",
	}
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Ping demo server",
	Long:  `Send a listUsers request then tries to list workspaces for each users on demo server`,
	Run: func(cmd *cobra.Command, args []string) {

		//check for the flags
		if protocol == "" {
			log.Fatal("Provide the protocol type")
		}
		if host == "" {
			log.Fatal("Provide the host")
		}
		if id == "" {
			log.Fatal("Provide the id")
		}
		if user == "" {
			log.Fatal("Provide the user")
		}
		if pwd == "" {
			log.Fatal("Provide the password")
		}
		if secret == "" {
			log.Fatal("Provide a secert key")
		}

		//connect to the api
		sdkConfig := &cells_sdk.SdkConfig{
			Url:          protocol + "://" + host,
			ClientKey:    id,
			ClientSecret: secret,
			User:         user,
			Password:     pwd,
			SkipVerify:   skipVerify,
		}
		httpClient := transport.GetHttpClient(sdkConfig)
		ctx, transport, err := transport.GetRestClientTransport(sdkConfig, false)
		if err != nil {
			log.Fatal(err)
		}
		apiClient := client.New(transport, strfmt.Default)

		// list users
		param := &user_service.SearchUsersParams{
			Context:    ctx,
			HTTPClient: httpClient,
		}

		result, err := apiClient.UserService.SearchUsers(param)
		if err != nil {
			fmt.Printf("could not list users: %s\n", err.Error())
			log.Fatal(err)
		}
		if len(result.Payload.Users) == 0 {
			er := fmt.Errorf("no users at all on this instance")
			log.Fatal(er)
		}
		var foundOne bool
		fmt.Printf("Found %d users on this Cells\n", len(result.Payload.Users))
		if len(result.Payload.Users) > 0 {
			for i, u := range result.Payload.Users {
				fmt.Println(i+1, " *********  ", u.Login)
			}
		}
		for u, p := range knownPwd {
			fmt.Println(" ----------------", u, "----------------")
			if e := listingUserFiles(u, p); e == nil {
				foundOne = true
			}
		}
		if !foundOne {
			log.Fatal("Could not find any workspaces for any users, something strange happened!")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Execution error", err)
	}

}

func init() {

	//7
	rootCmd.PersistentFlags().StringVarP(&protocol, "protocol", "t", "", "protocol type HTTP or HTTPS")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "a", "", "hostname of the cells instance")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "username")
	rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "", "password of the user")
	rootCmd.PersistentFlags().StringVarP(&id, "clientKey", "k", "", "put the clientKey found in pydio.json")
	rootCmd.PersistentFlags().StringVarP(&secret, "clientSecret", "s", "", "put the clientSecret found in pydio.json")

}

func listingUserFiles(login string, userPass string) error {

	uSdkConfig := &cells_sdk.SdkConfig{
		Url:          protocol + "://" + host,
		ClientKey:    id,
		ClientSecret: secret,
		User:         login,
		Password:     userPass,
		SkipVerify:   skipVerify,
	}

	uHttpClient := transport.GetHttpClient(uSdkConfig)
	ctx, t, err := transport.GetRestClientTransport(uSdkConfig, false)
	if err != nil {
		log.Fatal(err)
	}
	uApiClient := client.New(t, strfmt.Default)


	if err != nil {
		return fmt.Errorf("could not log in, not able to fetch the password for %s %s", login, err.Error())
	} else {
		log.Println("Successfully logged ", login)
	}

	params := &meta_service.GetBulkMetaParams{
		Body: &models.RestGetBulkMetaRequest{NodePaths: []string{
			"/*",
		}},
		Context:    ctx,
		HTTPClient: uHttpClient,
	}

	result, err := uApiClient.MetaService.GetBulkMeta(params)
	if err != nil {
		return fmt.Errorf("could not list meta %s", err.Error())
	}

	if len(result.Payload.Nodes) > 0 {
		fmt.Printf("* %d meta\n", len(result.Payload.Nodes))
		fmt.Println("USER ", login)

		for _, u := range result.Payload.Nodes {
			fmt.Println("  -", u.Path)

		}

	}

	return nil
}
