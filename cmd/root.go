package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	"github.com/pydio/cells-sdk-go"
	"github.com/pydio/cells-sdk-go/client"
	"github.com/pydio/cells-sdk-go/client/meta_service"
	"github.com/pydio/cells-sdk-go/client/user_service"
	"github.com/pydio/cells-sdk-go/models"
	"github.com/pydio/cells-sdk-go/transport"
	"github.com/pydio/cells-sdk-go/transport/http"
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
		"bob":   "bob",
		"alice": "alice",
	}
)

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Ping demo server",
	Long:  `This command sends a listUsers request to the demo server and then tries to list the workspaces for each of the default users`,
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
			log.Fatal("Provide a secret key")
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
		httpClient := http.GetHttpClient(sdkConfig)
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
			er := fmt.Errorf("no user at all on this instance")
			log.Fatal(er)
		}
		var foundOne bool
		fmt.Printf("Found %d users in this instance\n", len(result.Payload.Users))
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

	rootCmd.PersistentFlags().StringVarP(&protocol, "protocol", "t", "", "HTTP or HTTPS")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "a", "", "FQDN of this server")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "A registered admin user name")
	rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "", "A registered admin user password")
	rootCmd.PersistentFlags().StringVarP(&id, "clientKey", "k", "", "The front-end client key (can be found in the pydio.json)")
	rootCmd.PersistentFlags().StringVarP(&secret, "clientSecret", "s", "", "The front-end client secret (can be found in the pydio.json)")

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

	uHttpClient := http.GetHttpClient(uSdkConfig)
	ctx, t, err := transport.GetRestClientTransport(uSdkConfig, false)
	if err != nil {
		return fmt.Errorf("could not log in, not able to fetch the password for %s %s", login, err.Error())
	}
	uApiClient := client.New(t, strfmt.Default)

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
