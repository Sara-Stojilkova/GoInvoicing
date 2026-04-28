//go:build ignore

// Seed script: creates two test users via the Supabase Admin API,
// each placed in a pre-existing agency with the required signup metadata.
//
// Usage (from backend/):
//
//	SUPABASE_URL=... SUPABASE_SERVICE_ROLE_KEY=... go run ./scripts/seed_test_users.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type userMetadata struct {
	AgencyID string `json:"agency_id"`
	FullName string `json:"full_name"`
}

type appMetadata struct {
	AgencyID string `json:"agency_id"`
	Role     string `json:"role"`
}

type createUserRequest struct {
	Email        string       `json:"email"`
	Password     string       `json:"password"`
	EmailConfirm bool         `json:"email_confirm"`
	UserMetadata userMetadata `json:"user_metadata"`
	AppMetadata  appMetadata  `json:"app_metadata"`
}

var testUsers = []struct {
	agencyID string
	req      createUserRequest
}{
	{
		agencyID: "ad1c8a02-576c-4338-a53d-ed016d03d74c",
		req: createUserRequest{
			Email:        "alice@alpha.test",
			Password:     "password123",
			EmailConfirm: true,
			UserMetadata: userMetadata{
				AgencyID: "ad1c8a02-576c-4338-a53d-ed016d03d74c",
				FullName: "Alice Alpha",
			},
			AppMetadata: appMetadata{
				AgencyID: "ad1c8a02-576c-4338-a53d-ed016d03d74c",
				Role:     "member",
			},
		},
	},
	{
		agencyID: "6a3c2ade-9cca-4b40-bffa-c967d53778af",
		req: createUserRequest{
			Email:        "bob@beta.test",
			Password:     "password123",
			EmailConfirm: true,
			UserMetadata: userMetadata{
				AgencyID: "6a3c2ade-9cca-4b40-bffa-c967d53778af",
				FullName: "Bob Beta",
			},
			AppMetadata: appMetadata{
				AgencyID: "6a3c2ade-9cca-4b40-bffa-c967d53778af",
				Role:     "member",
			},
		},
	},
}

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || serviceRoleKey == "" {
		fmt.Fprintln(os.Stderr, "SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
		os.Exit(1)
	}

	client := &http.Client{}

	for _, u := range testUsers {
		body, _ := json.Marshal(u.req)

		req, err := http.NewRequest(
			http.MethodPost,
			supabaseURL+"/auth/v1/admin/users",
			bytes.NewReader(body),
		)
		if err != nil {
			fatalf("request error: %v\n", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
		req.Header.Set("apikey", serviceRoleKey)

		resp, err := client.Do(req)
		if err != nil {
			fatalf("http error: %v\n", err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			fatalf("create %s (status %d): %s\n", u.req.Email, resp.StatusCode, respBody)
		}

		var result struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil || result.ID == "" {
			fatalf("parse user id: %s\n", respBody)
		}
		fmt.Printf("created  %s → %s (agency %s)\n", u.req.Email, result.ID, u.agencyID)

		// Activate the user — trigger inserts with activated = false by default
		patchBody, _ := json.Marshal(map[string]any{"activated": true})
		patchReq, _ := http.NewRequest(
			http.MethodPatch,
			supabaseURL+"/rest/v1/users?id=eq."+result.ID,
			bytes.NewReader(patchBody),
		)
		patchReq.Header.Set("Content-Type", "application/json")
		patchReq.Header.Set("Authorization", "Bearer "+serviceRoleKey)
		patchReq.Header.Set("apikey", serviceRoleKey)
		patchReq.Header.Set("Prefer", "return=minimal")

		patchResp, err := client.Do(patchReq)
		if err != nil {
			fatalf("activate %s: %v\n", u.req.Email, err)
		}
		defer patchResp.Body.Close()

		if patchResp.StatusCode != http.StatusNoContent && patchResp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(patchResp.Body)
			fatalf("activate %s (status %d): %s\n", u.req.Email, patchResp.StatusCode, b)
		}
		fmt.Printf("activated %s\n\n", u.req.Email)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
