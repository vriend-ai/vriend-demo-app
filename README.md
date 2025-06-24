## Vriend Demo Application

This demonstration application serves as an introductory guide to an application that
integrates with the Vriend platform.

By following the straightforward instructions provided below, you can develop a fully
functional application that enables both you and your users to log in to Vriend using
SSO with a Vriend ID. Additionally, you will be able to make your initial API calls
against the platform.

### Quick Start

1. **Visit https://app.corp.vriend.ai and log in with your Vriend ID account.** If you do not have one, click on “Login” and register a new account.
2. **After successful login, navigate to https://app.corp.vriend.ai/organizations** to create your first Organization. This could be a team, department, or any other group.
3. **Navigate to https://app.corp.vriend.ai/applications** to create your first OIDC/OAuth2 application. For testing purposes, you can use a redirect URI with localhost, such as http://127.0.0.1:1234/oauth/redirect.
4. **Copy the generated .env file into the root directory** of this project and start the application with `go run main.go`.

### Next Steps

Refer to Vriend’s Open API documentation (https://api.corp.vriend.ai/docs) to obtain
a list of integration-ready API endpoints and detailed information about the business
logic and overall functionality.
