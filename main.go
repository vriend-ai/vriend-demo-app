package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
)

var (
	clientID     string
	clientSecret string
	redirectURL  string

	authTokenURL         string
	vriendAPIURL         string
	port                 = ":1234"
	redirectEndpointPath = "/oauth/redirect"

	cookieName = "access_token"
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	clientID = os.Getenv("VRIEND_CLIENT_ID")
	clientSecret = os.Getenv("VRIEND_CLIENT_SECRET")
	redirectURL = os.Getenv("APPLICATION_REDIRECT_URL")
	authTokenURL = os.Getenv("VRIEND_AUTH_TOKEN_URL")
	vriendAPIURL = os.Getenv("VRIEND_API_URL")
	if clientID == "" || clientSecret == "" || authTokenURL == "" || redirectURL == "" || vriendAPIURL == "" {
		log.Fatal("Environment variables VRIEND_CLIENT_ID, VRIEND_CLIENT_SECRET, VRIEND_AUTH_TOKEN_URL, and APPLICATION_REDIRECT_URL must be set")
	}

	// Extract the port (default 80 for http and 443 for https) from the redirect URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		log.Fatal("Invalid redirect URL: " + err.Error())
	}

	extractedPort := parsedURL.Port()
	if extractedPort != "" {
		port = ":" + extractedPort
	} else {
		// Default ports based on scheme
		if parsedURL.Scheme == "https" {
			port = ":443"
		} else {
			port = ":80"
		}
	}

	redirectEndpointPath = parsedURL.Path
}

func main() {
	// Serve static files from the public directory
	http.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	http.HandleFunc("GET "+redirectEndpointPath, handleOAuthCallback)
	http.HandleFunc("GET /logout", handleLogout)
	http.HandleFunc("GET /", handleIndex)

	log.Printf("Listening on %s (http://127.0.0.1%s)", port, port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// handleOAuthCallback handles the OAuth callback and exchanges the code for an access token.
func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)

	resp, err := http.Post(
		authTokenURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Set the token in a cookie if successful
	if resp.StatusCode == http.StatusOK {
		// Parse the token from the response body (assuming JSON with "access_token")
		type tokenResponse struct {
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			ExpiresIn    int    `json:"expires_in"`
			Scope        string `json:"scope"`
			RefreshToken string `json:"refresh_token,omitempty"`
		}
		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err == nil && tr.AccessToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    tr.AccessToken,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteLaxMode,
			})
		}
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Token exchange failed: "+string(body), resp.StatusCode)
		return
	}
	// Redirect to /
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
		Value:   "",
		Expires: time.Now().Add(-time.Second), // Set an expiration in the past to delete the cookie
	})
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusSeeOther)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Construct the OAuth login URL
	authURL := "http://127.0.0.1:9100/application/o/authorize/?" + url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {"openid profile email vriend_stresslevel"}, // adjust scope as needed
	}.Encode()

	authURLWithoutScope := "http://127.0.0.1:9100/application/o/authorize/?" + url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {"openid profile email"}, // adjust scope as needed
	}.Encode()

	// Parse and execute the template
	tmpl, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		LoginURL             string
		VriendAPIURL         string
		LoginURLWithoutScope string
	}{
		LoginURL:             authURL,
		VriendAPIURL:         vriendAPIURL,
		LoginURLWithoutScope: authURLWithoutScope,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
