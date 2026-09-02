package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	insecurecreds "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	v1 "github.com/firebird-dc/firebird-proto-contracts/gen/go"
)

func parseTarget(apiURL string, insecure bool) (string, bool, error) {
	if strings.Contains(apiURL, "://") {
		u, err := url.Parse(apiURL)
		if err != nil {
			return "", false, fmt.Errorf("parse api-url: %w", err)
		}
		if u.Hostname() == "" {
			return "", false, fmt.Errorf("invalid api-url: %s", apiURL)
		}

		port := u.Port()
		if port == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return "", false, fmt.Errorf("unsupported scheme: %s", u.Scheme)
			}
		}

		useTLS := u.Scheme == "https" && !insecure
		return fmt.Sprintf("%s:%s", u.Hostname(), port), useTLS, nil
	}

	return apiURL, !insecure, nil
}

func main() {
	var (
		apiURL       = flag.String("api-url", "", "gRPC host:port or URL, e.g. https://api.example.com or localhost:9090")
		clientID     = flag.String("client-id", "", "service account client_id")
		clientSecret = flag.String("client-secret", "", "service account client_secret")
		timeout      = flag.Duration("timeout", 15*time.Second, "per-request timeout")
		insecure     = flag.Bool("insecure", false, "use insecure gRPC connection")
	)
	flag.Parse()

	if *apiURL == "" || *clientID == "" || *clientSecret == "" {
		flag.Usage()
		log.Fatal("api-url, client-id and client-secret are required")
	}

	target, useTLS, err := parseTarget(*apiURL, *insecure)
	if err != nil {
		log.Fatalf("invalid api-url: %v", err)
	}

	log.Printf("connecting to %s (tls=%v)", target, useTLS)

	dialCtx, cancelDial := context.WithTimeout(context.Background(), *timeout)
	defer cancelDial()

	var transportCreds credentials.TransportCredentials
	if useTLS {
		transportCreds = credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})
	} else {
		transportCreds = insecurecreds.NewCredentials()
	}

	conn, err := grpc.DialContext(
		dialCtx,
		target,
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("dial gRPC: %v", err)
	}
	defer conn.Close()

	client := v1.NewUserManagementServiceClient(conn)

	// 1) Token
	tokenCtx, cancelToken := context.WithTimeout(context.Background(), *timeout)
	defer cancelToken()

	tokenResp, err := client.Token(tokenCtx, &v1.TokenRequest{
		GrantType:    "client_credentials",
		ClientId:     *clientID,
		ClientSecret: *clientSecret,
	})
	if err != nil {
		log.Fatalf("Token RPC failed: %v", err)
	}
	if tokenResp.GetAccessToken() == "" {
		log.Fatal("Token RPC returned empty access_token")
	}

	log.Printf("token acquired successfully, type=%s expires_in=%d access_token_prefix=%s...",
		tokenResp.GetTokenType(),
		tokenResp.GetExpiresIn(),
		prefix(tokenResp.GetAccessToken(), 24),
	)

	// 2) GetUserInfo with bearer token
	meCtx, cancelMe := context.WithTimeout(context.Background(), *timeout)
	defer cancelMe()

	meCtx = metadata.NewOutgoingContext(
		meCtx,
		metadata.Pairs("authorization", "Bearer "+tokenResp.GetAccessToken()),
	)

	meResp, err := client.GetUserInfo(meCtx, &v1.GetUserInfoRequest{})
	if err != nil {
		log.Fatalf("GetUserInfo RPC failed: %v", err)
	}

	out, err := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}.Marshal(meResp)
	if err != nil {
		log.Fatalf("marshal GetUserInfo response: %v", err)
	}

	fmt.Println("Current user info:")
	fmt.Println(string(out))
}

func prefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
